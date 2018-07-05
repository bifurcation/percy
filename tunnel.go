package percy

import (
	"log"
	"net"

	"github.com/bifurcation/mint/syntax"
)

type ProtectionProfile uint16

type HBHKeys struct {
	Marker         uint8
	Profile        uint16
	ClientWriteKey []byte `tls:"head=1"`
	ServerWriteKey []byte `tls:"head=1"`
	MasterSalt     []byte `tls:"head=1"`
}

type KMFTunnel interface {
	Send(assoc AssociationID, msg []byte) error
}

type MDDTunnel interface {
	Send(assoc AssociationID, msg []byte) error
	SetKeys(assocID AssociationID, keys HBHKeys) error
}

//////////

const (
	kdBufferSize = 2048
)

type UDPForwarder struct {
	MD     MDDTunnel
	server *net.UDPAddr
	conns  map[AssociationID]*net.UDPConn
}

func NewUDPForwarder(server string) (*UDPForwarder, error) {
	serverAddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		return nil, err
	}

	return &UDPForwarder{
		server: serverAddr,
		conns:  map[AssociationID]*net.UDPConn{},
	}, nil
}

func (fwd *UDPForwarder) monitor(assocID AssociationID, conn *net.UDPConn) {
	buf := make([]byte, kdBufferSize)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading KD socket: %v", err)
			return
		}
		buf = buf[:n]

		log.Printf("MD <-- KD for %v with [%d] bytes", assocID, len(buf))

		switch packetClass(buf) {
		case packetClassDTLS:
			err = fwd.MD.Send(assocID, buf)
			if err != nil {
				log.Printf("Error forwarding DTLS packet: %v", err)
			}

		case packetClassHBHKey:
			var keys HBHKeys
			_, err := syntax.Unmarshal(buf, &keys)
			if err != nil {
				log.Printf("Error parsing HBHKeys struct: %v", err)
			}

			fwd.MD.SetKeys(assocID, keys)
		}

		buf = buf[:kdBufferSize]
	}
}

func (fwd *UDPForwarder) Send(assocID AssociationID, msg []byte) error {
	var err error
	conn, ok := fwd.conns[assocID]
	if !ok {
		conn, err = net.DialUDP("udp", nil, fwd.server)
		if err != nil {
			return err
		}

		conn.SetReadBuffer(kdBufferSize)

		fwd.conns[assocID] = conn
		go fwd.monitor(assocID, conn)
	}

	log.Printf("MD --> KD for %v with [%d] bytes", assocID, len(msg))

	_, err = conn.Write(msg)
	return err
}
