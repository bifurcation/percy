package percy

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"
)

type Client struct {
	RecvChan chan []byte

	name       string
	clientAddr *net.UDPAddr
	serverAddr *net.UDPAddr
	serverConn *net.UDPConn
	stopChan   chan bool
}

func NewClient(name string, clientPort, serverPort int) (*Client, error) {
	c := new(Client)
	c.RecvChan = make(chan []byte, 10)
	c.name = name
	c.clientAddr = &net.UDPAddr{Port: clientPort}
	c.serverAddr = &net.UDPAddr{Port: serverPort}

	var err error
	c.serverConn, err = net.DialUDP("udp", c.clientAddr, c.serverAddr)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Listen() error {
	go func(c *Client) {
		buf := make([]byte, 2048)

		for {
			select {
			case <-c.stopChan:
				log.Println("Stopping Client")
			default:
			}

			n, addr, err := c.serverConn.ReadFromUDP(buf)

			if err != nil {
				log.Printf("Recv Error: %v\n", err)
				continue
			} else {
				log.Printf("Recv [%s] [%d] [%s] [%s]", c.name, n, addr.String(), string(buf[:n]))
				c.RecvChan <- buf[:n]
			}
		}
	}(c)

	return nil
}

func (c *Client) Write(msg []byte) error {
	_, err := c.serverConn.Write(msg)
	return err
}

func (c *Client) Stop() {
	c.stopChan <- true
}

/////

func caller() string {
	_, file, line, _ := runtime.Caller(2)
	splits := strings.Split(file, "/")
	filename := splits[len(splits)-1]
	return fmt.Sprintf("%s:%d:", filename, line)
}

func Assert(t *testing.T, result bool, message string) {
	if !result {
		t.Fatalf("%s %s", caller(), message)
	}
}

func AssertNotError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s %s: %s", caller(), message, err)
	}
}

const packetTimeout = 100 * time.Millisecond

func AssertRecvPacket(t *testing.T, c *Client, msg []byte, message string) {
	select {
	case packet := <-c.RecvChan:
		if !bytes.Equal(msg, packet) {
			t.Fatalf("%s %s: [%x] != [%x]", caller(), message, msg, packet)
		}
	case <-time.After(packetTimeout):
		t.Fatalf("%s %s: %s", caller(), message, "timeout")
	}
}

/////

func TestRTPForwarding(t *testing.T) {
	// Set up the server
	mdd := NewMDD()
	err := mdd.Listen(8888)
	AssertNotError(t, err, "Error creating MDD")

	// Test RTP forwarding
	c1, err := NewClient("c1", 9999, 8888)
	AssertNotError(t, err, "Error creating client 1")
	c1.Listen()

	c2, err := NewClient("c2", 7777, 8888)
	AssertNotError(t, err, "Error creating client 2")
	c2.Listen()

	m1 := []byte("c1-init")
	m2 := []byte("c2-probe")
	m3 := []byte("c1-probe")

	c1.Write(m1) // Primes c1 as a recipient
	c2.Write(m2)
	AssertRecvPacket(t, c1, m2, "Failed to forward c2->c1")
	c1.Write(m3)
	AssertRecvPacket(t, c2, m3, "Failed to forward c1->c2")

}
