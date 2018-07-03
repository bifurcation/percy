package percy

import (
	"log"
	"net"
)

type ProtectionProfile uint16

type SRTPKeys struct {
	MasterKeyID []byte
	ClientKey   []byte
	ServerKey   []byte
	ClientSalt  []byte
	ServerSalt  []byte
}

type KMFTunnel interface {
	Send(assoc AssociationID, msg []byte) error
	SendWithProfiles(assocID AssociationID, msg []byte, profiles []ProtectionProfile) error
}

type MDDTunnel interface {
	Send(assoc AssociationID, msg []byte) error
	SendWithKeys(assocID AssociationID, msg []byte, profile ProtectionProfile, keys SRTPKeys) error
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

		err = fwd.MD.Send(assocID, buf)
		if err != nil {
			log.Printf("Error forwarding DTLS packet: %v", err)
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

func (fwd *UDPForwarder) SendWithProfiles(assocID AssociationID, msg []byte, profiles []ProtectionProfile) error {
	// TODO Do something with the profiles
	return fwd.Send(assocID, msg)
}
