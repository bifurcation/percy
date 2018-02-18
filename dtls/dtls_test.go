package dtls

import (
	"bytes"
	"testing"
)

const (
	keyFile     = "key.pem"
	certFile    = "cert.pem"
	maxRTT      = 20
	srtpKeySize = 60
)

func TestDTLS(t *testing.T) {
	// Initialize client and server instances
	client, err := NewDTLSClient(keyFile, certFile)
	defer client.Close()
	if err != nil {
		t.Fatalf("Error creating DTLS client: %v", err)
	}

	server, err := NewDTLSServer(keyFile, certFile)
	defer server.Close()
	if err != nil {
		t.Fatalf("Error creating DTLS server: %v", err)
	}

	// Broker a DTLS handshake
	rtt := 0
	for !client.Done() && !server.Done() && rtt < maxRTT {
		rtt += 1

		client.Kick()

		packet := client.Recv()
		for len(packet) > 0 {
			server.Send(packet)
			packet = client.Recv()
		}

		server.Kick()
		packet = server.Recv()
		for len(packet) > 0 {
			client.Send(packet)
			packet = server.Recv()
		}
	}

	// Verify that it succeeded
	if rtt == maxRTT {
		t.Fatalf("Handshake failed to converge")
	}

	// Verify that we get matching SRTP parameters
	clientProfile := client.SRTPProfile()
	serverProfile := server.SRTPProfile()
	if clientProfile != serverProfile {
		t.Fatalf("SRTP profile mismatch [%04x] != [%04x]", clientProfile, serverProfile)
	}

	clientKey := client.SRTPKey(srtpKeySize)
	serverKey := server.SRTPKey(srtpKeySize)
	if !bytes.Equal(clientKey, serverKey) {
		t.Fatalf("SRTP key mismatch [%x] != [%x]", clientKey, serverKey)
	}
}
