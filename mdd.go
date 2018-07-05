package percy

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/fluffy/rtp"
)

type AssociationID uint16

type dtlsSRTPPacketClass uint8

const (
	packetClassDTLS dtlsSRTPPacketClass = iota
	packetClassSRTP
	packetClassSTUN
	packetClassHBHKey
	packetClassUnknown
)

// https://tools.ietf.org/html/rfc5764#section-5.1.2
func packetClass(msg []byte) dtlsSRTPPacketClass {
	if len(msg) == 0 {
		return packetClassUnknown
	}

	// XXX: We could do more validation that DTLS and SRTP
	//      packets are well-formed
	B := msg[0]
	switch {
	case 127 < B && B < 192:
		return packetClassSRTP
	case 19 < B && B < 64:
		return packetClassDTLS
	case B < 2:
		return packetClassSTUN
	case B == 0xFF:
		return packetClassHBHKey
	default:
		return packetClassUnknown
	}
}

type packet struct {
	addr *net.UDPAddr
	msg  []byte
}

func addrToAssoc(addr *net.UDPAddr) AssociationID {
	h := sha256.New()
	h.Write([]byte(addr.String()))
	sum := h.Sum(nil)
	return AssociationID((uint16(sum[0]) << 8) + uint16(sum[1]))
}

type MDD struct {
	name        string
	addr        *net.UDPAddr
	conn        *net.UDPConn
	clients     map[AssociationID]*net.UDPAddr
	rtpSessions map[AssociationID]*rtp.RTPSession
	stopChan    chan bool
	doneChan    chan bool
	packetChan  chan packet
	timeout     time.Duration

	KD       KMFTunnel
	keys     map[AssociationID]HBHKeys
	profile  ProtectionProfile
	profiles []ProtectionProfile
	// TODO add some mutexes
}

func NewMDD() *MDD {
	mdd := new(MDD)
	mdd.name = "mdd"
	mdd.clients = map[AssociationID]*net.UDPAddr{}
	mdd.rtpSessions = map[AssociationID]*rtp.RTPSession{}
	mdd.timeout = 10 * time.Millisecond

	mdd.stopChan = make(chan bool)
	mdd.doneChan = make(chan bool)
	mdd.packetChan = make(chan packet)

	// TODO Add some defaults
	mdd.profiles = []ProtectionProfile{}
	mdd.keys = map[AssociationID]HBHKeys{}

	return mdd
}

func (mdd *MDD) handleDTLS(assocID AssociationID, msg []byte) {
	// TODO Notify the KD of supported SRTP profiles
	mdd.KD.Send(assocID, msg)
}

func (mdd *MDD) handleHBHKey(assocID AssociationID, msg []byte) {
	log.Printf("Received HBH key from KMF: %v", msg)
}

func (mdd *MDD) broadcast(assocID AssociationID, msg []byte) {
	// Send the packet out to all the clients except
	// the one that sent it
	for client, addr := range mdd.clients {
		if client == assocID {
			continue
		}

		log.Printf("Client <-- MD for %v[%v] with [%d] bytes", client, addr, len(msg))

		_, err := mdd.conn.WriteToUDP(msg, addr)
		if err != nil {
			log.Println("Error forwarding packet")
		}
	}
}

func (mdd *MDD) handleSTUN(addr *net.UDPAddr, msg []byte) {
	message, err := ParseSTUN(msg)
	if err != nil {
		log.Println("Error parsing STUN message", err, msg)
		return
	}

	log.Println(addr, message.header)

	switch message.msgType {
	case MSG_TYPE_REQUEST:
		response := STUNMessage{header: message.header}
		switch message.header.Type {
		case MSG_BINDING:
			response.msgType = MSG_TYPE_SUCCESS
			// 22 to 256 alphanumeric characters
			response.icePassword = "abcdefabcdefabcdefabcdefabcdefab"
			response.AddXorMappedAddress(addr)
			response.AddMessageIntegrity()
			response.AddFingerprint()
		default:
			log.Printf("Unhandled STUN message type: %v", message)
			response.msgType = MSG_TYPE_ERROR
			response.AddErrorCode(500, "Unimplemented")
		}

		responseBytes, err := response.Serialize()
		if err != nil {
			log.Println("Error serializing response:", err)
			return
		}
		log.Println("Sending", response.header)

		_, err = mdd.conn.WriteToUDP(responseBytes, addr)
		if err != nil {
			log.Println("Error replying to STUN request:", err)
		}
	case MSG_TYPE_INDICATION:
		// TODO: handle received indications
	case MSG_TYPE_SUCCESS:
		// TODO: handle received responses
	case MSG_TYPE_ERROR:
		// TODO: handle received errors
	}
}

func (mdd *MDD) Listen(port int) error {
	var err error

	mdd.addr = &net.UDPAddr{Port: port}
	mdd.conn, err = net.ListenUDP("udp", mdd.addr)
	if err != nil {
		return err
	}

	mdd.packetChan = make(chan packet, 10)

	go func(packetChan chan packet) {
		buf := make([]byte, 2048)

		for {
			n, addr, err := mdd.conn.ReadFromUDP(buf)

			if err == nil {
				packetChan <- packet{addr: addr, msg: buf[:n]}
			}
			// TODO log errors
		}
	}(mdd.packetChan)

	go func(mdd *MDD) {
		for {
			var pkt packet

			select {
			case <-mdd.stopChan:
				mdd.doneChan <- true
				return
			case <-time.After(mdd.timeout):
				continue
			case pkt = <-mdd.packetChan:
			}

			if err != nil {
				log.Printf("Recv Error: %v", err)
				continue
			}

			assocID := addrToAssoc(pkt.addr)

			log.Printf("Client --> MD for %v[%v] with [%d] bytes", assocID, pkt.addr, len(pkt.msg))

			// Remember the client if it's new
			// XXX: Could have an interface to add/remove clients, then
			//      just filter unknown clients here.
			if _, ok := mdd.clients[assocID]; !ok {
				mdd.clients[assocID] = pkt.addr
				// setup a RTP Session handle
				mdd.rtpSessions[assocID] = rtp.NewRTPSession()
			}

			// XXX: For now, all packets are re-broadcast, which means
			// this will only really work in cases where there are only
			// two clients.
			//
			// XXX: DTLS packets can be routed to a local DTLS stack as
			// soon as we have one, and can get the keys out to
			// re-encrypt.
			//
			// XXX: Handling STUN locally will require routing SDP
			// offer/answer via the MD, so that it can grab the ICE ufrag
			// and password and use them to synthesize STUN responses.
			switch packetClass(pkt.msg) {
			case packetClassDTLS:
				mdd.handleDTLS(assocID, pkt.msg)
			case packetClassSRTP:
				mdd.broadcast(assocID, pkt.msg)
			case packetClassSTUN:
				mdd.handleSTUN(pkt.addr, pkt.msg)
			case packetClassHBHKey:
				mdd.handleHBHKey(assocID, pkt.msg)
			default:
				log.Printf("Unknown packet type received")
			}
		}
	}(mdd)

	return nil
}

func (mdd *MDD) Send(assocID AssociationID, msg []byte) error {
	addr, ok := mdd.clients[assocID]
	log.Printf("Client <-- MD for %v[%v] with [%d] bytes", assocID, addr, len(msg))
	if !ok {
		return fmt.Errorf("Unknown client [%04x]", assocID)
	}

	_, err := mdd.conn.WriteToUDP(msg, addr)
	return err
}

func (mdd *MDD) SetKeys(assocID AssociationID, keys HBHKeys) error {
	log.Printf("       --- MD setting keys for %v", assocID)
	mdd.keys[assocID] = keys
	return nil
}

func (mdd *MDD) Stop() {
	mdd.stopChan <- true
	<-mdd.doneChan

	mdd.conn.Close()

	// Avoid race conditions
	<-time.After(10 * time.Millisecond)
}
