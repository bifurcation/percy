package percy

import (
	"bytes"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Client struct {
	RecvChan chan []byte

	name       string
	clientAddr *net.UDPAddr
	serverAddr *net.UDPAddr
	conn       *net.UDPConn
	stopChan   chan bool
	doneChan   chan bool
	packetChan chan packet
}

func NewClient(name string, clientPort, serverPort int) (*Client, error) {
	c := new(Client)
	c.RecvChan = make(chan []byte, 10)
	c.name = name
	c.clientAddr = &net.UDPAddr{Port: clientPort}
	c.serverAddr = &net.UDPAddr{Port: serverPort}

	c.stopChan = make(chan bool, 1)
	c.doneChan = make(chan bool, 1)

	var err error
	c.conn, err = net.DialUDP("udp", c.clientAddr, c.serverAddr)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Listen() error {

	c.packetChan = make(chan packet, 10)

	go func(packetChan chan packet) {
		buf := make([]byte, 2048)

		for {
			n, addr, err := c.conn.ReadFromUDP(buf)

			if err == nil {
				packetChan <- packet{addr: addr, msg: buf[:n]}
			}
			// TODO log errors
		}
	}(c.packetChan)

	go func(c *Client) {
		for {
			var pkt packet

			select {
			case <-c.stopChan:
				c.doneChan <- true
				return
			case pkt = <-c.packetChan:
				c.RecvChan <- pkt.msg
			case <-time.After(pause):
			}
		}
	}(c)

	return nil
}

func (c *Client) Write(msg []byte) error {
	_, err := c.conn.Write(msg)
	return err
}

func (c *Client) Stop() {
	c.stopChan <- true
	<-c.doneChan
	c.conn.Close()
}

type CountTunnel struct {
	sendCount             int
	sendWithProfilesCount int
}

func (tun *CountTunnel) Send(assoc AssociationID, msg []byte) error {
	tun.sendCount += 1
	return nil
}

func (tun *CountTunnel) SendWithProfiles(assoc AssociationID, msg []byte, profiles []ProtectionProfile) error {
	tun.sendWithProfilesCount += 1
	return nil
}

/////

func caller() string {
	_, file, line, _ := runtime.Caller(2)
	splits := strings.Split(file, "/")
	filename := splits[len(splits)-1]
	return fmt.Sprintf("%s:%d:", filename, line)
}

func AssertNotNil(t *testing.T, obj interface{}, message string) {
	if obj == nil {
		t.Fatalf("%s %s", caller(), message)
	}
}

func AssertNotError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s %s: %s", caller(), message, err)
	}
}

func AssertEquals(t *testing.T, one interface{}, two interface{}) {
	if one != two {
		t.Fatalf("%s [%v] != [%v]", caller(), one, two)
	}
}

const packetTimeout = 1000 * time.Millisecond

func AssertRecvPacket(t *testing.T, c *Client, msg []byte, message string) {
	select {
	case packet := <-c.RecvChan:
		if cmp := bytes.Compare(msg, packet); cmp != 0 {
			t.Fatalf("%s %s: [%x] != [%x] ~ [%d]", caller(), message, msg, packet, cmp)
		}
	case <-time.After(packetTimeout):
		t.Fatalf("%s %s: %s", caller(), message, "timeout")
	}
}

func AssertNotRecvPacket(t *testing.T, c *Client, message string) {
	select {
	case <-c.RecvChan:
		t.Fatalf("%s %s: %s", caller(), "Should not have received a packet")
	case <-time.After(packetTimeout):
	}
}

/////

var (
	// XXX: If you make nClients too big, you run into race
	// conditions with the n^2 tests
	nClients       = 10
	pause          = 10 * time.Millisecond
	serverPort     = 8888
	clientBasePort = 9999
)

func TestRTPForwarding(t *testing.T) {
	// Set up the server
	mdd := NewMDD(&CountTunnel{})
	err := mdd.Listen(serverPort)
	AssertNotError(t, err, "Error creating MDD")
	defer mdd.Stop()

	// Start up a bunch of clients
	clients := make([]*Client, nClients)
	for i := range clients {
		port := 9999 + i
		clients[i], err = NewClient("c"+strconv.Itoa(i), port, serverPort)
		AssertNotError(t, err, fmt.Sprintf("Error creating client %d", i))
		clients[i].Listen()
		defer clients[i].Stop()
	}

	// Verify that previous clients hear new clients when they join
	for i, sender := range clients {
		srtpPacket := []byte{128, 0, byte(i)}
		sender.Write(srtpPacket)

		for j := 0; j < i; j += 1 {
			AssertRecvPacket(t, clients[j], srtpPacket,
				fmt.Sprintf("Failed to forward c%d->c%d", i, j))
		}
	}

	// Verify that packets from joined clients broadcast to everyone
	// but the sender
	for i, sender := range clients {
		srtpPacket := []byte{128, 1, byte(i)}
		sender.Write(srtpPacket)

		<-time.After(10 * time.Millisecond)

		for j := 0; j < i; j += 1 {
			if i == j {
				AssertNotRecvPacket(t, clients[j], "Loopback detected")
				continue
			}

			AssertRecvPacket(t, clients[j], srtpPacket,
				fmt.Sprintf("Failed to broadcast c%d->c%d", i, j))
		}
	}
}

func TestDTLSForwarding(t *testing.T) {
	// Set up the server and a client
	tun := &CountTunnel{sendCount: 0, sendWithProfilesCount: 0}
	mdd := NewMDD(tun)
	err := mdd.Listen(serverPort)
	AssertNotError(t, err, "Error creating MDD")
	defer mdd.Stop()

	client, err := NewClient("cDTLS", clientBasePort, serverPort)
	AssertNotError(t, err, "Error creating Client")
	client.Listen()
	defer client.Stop()

	// Test forward direction (ClientHello)
	dtlsPacket := bytes.Repeat([]byte{0}, 14)
	dtlsPacket[0] = 0x16
	dtlsPacket[13] = 0x01
	err = client.Write(dtlsPacket)
	AssertNotError(t, err, "Error sending DTLS packet")
	<-time.After(pause)
	AssertEquals(t, tun.sendWithProfilesCount, 1)

	// Test forward direction (non-ClientHello)
	dtlsPacket = []byte{0x16, 0x00}
	err = client.Write(dtlsPacket)
	AssertNotError(t, err, "Error sending DTLS packet")
	<-time.After(pause)
	AssertEquals(t, tun.sendCount, 1)

	// Grab the association ID for our client
	var assoc AssociationID
	for clientID := range mdd.clients {
		assoc, err = stringToAssoc(clientID)
		AssertNotError(t, err, "MDD provisioned an invalid client ID")
	}

	// Test reverse direction (no keys)
	dtlsPacket = []byte{0x16, 0x01}
	mdd.Send(assoc, dtlsPacket)
	AssertRecvPacket(t, client, dtlsPacket, "DTLS packet was not forwarded")

	// Test reverse direction (keys)
	dtlsPacket = []byte{0x16, 0x02}
	keys := SRTPKeys{}
	profile := ProtectionProfile(0xff)
	mdd.SendWithKeys(assoc, dtlsPacket, profile, keys)
	AssertRecvPacket(t, client, dtlsPacket, "DTLS packet was not forwarded with keys")
	AssertEquals(t, mdd.profile, profile)
	AssertNotNil(t, mdd.keys, "MDD failed to set keys")
}

func TestDiscrimination(t *testing.T) {
	// Set up the server
	tun := &CountTunnel{sendCount: 0, sendWithProfilesCount: 0}
	mdd := NewMDD(tun)
	err := mdd.Listen(serverPort)
	AssertNotError(t, err, "Error creating MDD")
	defer mdd.Stop()

	// Set up first client
	client1, err := NewClient("cDTLS", clientBasePort, serverPort)
	AssertNotError(t, err, "Error creating Client")
	client1.Listen()
	defer client1.Stop()

	// Have client send a ClientHello and verify that it goes through
	clientHelloPacket := bytes.Repeat([]byte{0}, 14)
	clientHelloPacket[0] = 0x16
	clientHelloPacket[13] = 0x01
	err = client1.Write(clientHelloPacket)
	<-time.After(pause)
	AssertEquals(t, tun.sendWithProfilesCount, 1)
	AssertEquals(t, tun.sendCount, 0)

	// Grab the association ID for client 1
	var client1ID string
	var client1Assoc AssociationID
	for clientID := range mdd.clients {
		client1ID = clientID
	}
	client1Assoc, err = stringToAssoc(client1ID)
	AssertNotError(t, err, "MDD provisioned an invalid client ID")

	// Set up second client
	client2, err := NewClient("cDTLS", clientBasePort+1, serverPort)
	AssertNotError(t, err, "Error creating Client")
	client2.Listen()
	defer client2.Stop()

	err = client2.Write(clientHelloPacket)
	<-time.After(pause)
	AssertEquals(t, tun.sendWithProfilesCount, 2)
	AssertEquals(t, tun.sendCount, 0)
	AssertNotRecvPacket(t, client1, "DTLS packet forwarded to other client")

	var client2ID string
	var client2Assoc AssociationID
	for clientID := range mdd.clients {
		if clientID != client1ID {
			client2ID = clientID
		}
	}
	client2Assoc, err = stringToAssoc(client2ID)
	AssertNotError(t, err, "MDD provisioned an invalid client ID")

	// Test RTP forwarding
	srtpPacket := []byte{128, 1, 0}
	client1.Write(srtpPacket)
	AssertRecvPacket(t, client2, srtpPacket, "SRTP packet not forwarded")
	AssertNotRecvPacket(t, client1, "SRTP packet forwarded to sender")

	// Test DTLS to client 1
	dtlsPacket := []byte{0x16, 0x01}
	mdd.Send(client1Assoc, dtlsPacket)
	AssertRecvPacket(t, client1, dtlsPacket, "DTLS packet was not forwarded")
	AssertNotRecvPacket(t, client2, "DTLS packet forwarded to other client")

	// Test DTLS to client 2
	dtlsPacket = []byte{0x16, 0x01}
	mdd.Send(client2Assoc, dtlsPacket)
	AssertRecvPacket(t, client2, dtlsPacket, "DTLS packet was not forwarded")
	AssertNotRecvPacket(t, client1, "DTLS packet forwarded to other client")
}
