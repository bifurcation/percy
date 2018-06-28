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

type UDPForwarder struct {
	mdd    MDDTunnel
	server *net.UDPAddr
	conns  map[AssociationID]*net.UDPConn
}

func NewUDPForwarder(mdd MDDTunnel, server string) (*UDPForwarder, error) {
	serverAddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		return nil, err
	}

	return &UDPForwarder{
		mdd:    mdd,
		server: serverAddr,
		conns:  map[AssociationID]*net.UDPConn{},
	}, nil
}

func (fwd *UDPForwarder) monitor(assocID AssociationID, conn *net.UDPConn) {
	buf := make([]byte, 2048)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading KD socket: %v", err)
			return
		}
		log.Printf("Forwarder <-- from AssocId %v with [%d] bytes", assocID, n)
		buf = buf[:n]
		err = fwd.mdd.Send(assocID, buf)
		if err != nil {
			log.Printf("Error forwarding DTLS packet: %v", err)
		}
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

		fwd.conns[assocID] = conn
		go fwd.monitor(assocID, conn)
	}

	_, err = conn.Write(msg)
	return err
}

func (fwd *UDPForwarder) SendWithProfiles(assocID AssociationID, msg []byte, profiles []ProtectionProfile) error {
	// TODO Do something with the profiles
	return fwd.Send(assocID, msg)
}

func (fwd *UDPForwarder) SendWithKeys(assocID AssociationID, msg []byte, profile ProtectionProfile, keys SRTPKeys) error {
	// TODO Do something with the profiles
	return fwd.mdd.SendWithKeys(assocID, msg, profile, keys)
}
