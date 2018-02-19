package dtls

import (
	"fmt"
	"testing"

	"github.com/bifurcation/percy/assert"
)

const (
	keyFile     = "../static/key.pem"
	certFile    = "../static/cert.pem"
	maxRTT      = 20
	srtpKeySize = 60
)

func TestDTLS(t *testing.T) {
	// Initialize client and server instances
	client, err := NewDTLSClient(keyFile, certFile)
	defer client.Close()
	assert.NotError(t, err, "Error creating DTLS client")

	server, err := NewDTLSServer(keyFile, certFile)
	defer server.Close()
	assert.NotError(t, err, "Error creating DTLS server")

	// Broker a DTLS handshake
	rtt := 0
	total := 0
	for !client.Done() && !server.Done() && rtt < maxRTT {
		rtt += 1
		fmt.Printf("rtt=%d ", rtt)

		total = 0
		client.Kick()
		packet := client.Recv()
		total += len(packet)
		for len(packet) > 0 {
			total += len(packet)
			server.Send(packet)
			packet = client.Recv()
		}
		fmt.Printf("c2s=%d ", total)

		total = 0
		server.Kick()
		packet = server.Recv()
		total += len(packet)
		for len(packet) > 0 {
			total += len(packet)
			client.Send(packet)
			packet = server.Recv()
		}
		fmt.Printf("s2c=%d \n", total)
	}

	// Verify that it succeeded
	assert.NotEqual(t, rtt, maxRTT, "Handshake failed to converge")

	// Verify that we get matching SRTP parameters
	clientProfile := client.SRTPProfile()
	serverProfile := server.SRTPProfile()
	assert.Equal(t, clientProfile, serverProfile, "SRTP profile mismatch")

	clientKey := client.SRTPKey(srtpKeySize)
	serverKey := server.SRTPKey(srtpKeySize)
	assert.BytesEqual(t, clientKey, serverKey, "SRTP key mismatch")
}
