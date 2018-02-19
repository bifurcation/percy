package percy

import (
	"testing"

	"github.com/bifurcation/percy/assert"
	"github.com/bifurcation/percy/dtls"
	"github.com/bifurcation/percy/srtp"
)

const (
	keyFile       = "./static/key.pem"
	certFile      = "./static/cert.pem"
	maxRTT        = 5
	ssrc1         = 0x01020304
	ssrc2         = 0x05060708
	testPacketLen = 30
)

type Endpoint struct {
	dtls    *dtls.DTLS
	srtpIn  *srtp.SRTP
	srtpOut *srtp.SRTP
}

func (e *Endpoint) Close() {
	e.dtls.Close()
	e.srtpIn.Close()
	e.srtpOut.Close()
}

func (e *Endpoint) SetupSRTP(t *testing.T, server bool) {
	profile := int(e.dtls.SRTPProfile())

	keyLen, err := srtp.KeyLength(profile)
	assert.NotError(t, err, "Error getting key length (1)")

	saltLen, err := srtp.SaltLength(profile)
	assert.NotError(t, err, "Error getting salt length (1)")

	key := e.dtls.SRTPKey(2*keyLen + 2*saltLen)

	clientMasterKey := key[:keyLen]
	key = key[keyLen:]

	serverMasterKey := key[:keyLen]
	key = key[keyLen:]

	clientMasterSalt := key[:saltLen]
	key = key[saltLen:]

	serverMasterSalt := key[:saltLen]
	key = key[saltLen:]

	clientSRTPKey := append(clientMasterKey, clientMasterSalt...)
	serverSRTPKey := append(serverMasterKey, serverMasterSalt...)

	inKey, outKey := serverSRTPKey, clientSRTPKey
	if server {
		inKey, outKey = clientSRTPKey, serverSRTPKey
	}

	e.srtpIn, err = srtp.NewSRTP(srtp.AnySSRCInbound, profile, inKey)
	assert.NotError(t, err, "Error setting up inbound SRTP")

	e.srtpOut, err = srtp.NewSRTP(srtp.AnySSRCOutbound, profile, outKey)
	assert.NotError(t, err, "Error setting up outbound SRTP")
}

func Handshake(t *testing.T, client, server *dtls.DTLS) {
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

	if rtt >= maxRTT {
		t.Fatalf("Max RTT exceeded: %d > %d", rtt, maxRTT)
	}
}

func TestDTLSSRTP(t *testing.T) {
	client1 := Endpoint{}
	client2 := Endpoint{}
	relay1 := Endpoint{}
	relay2 := Endpoint{}
	defer client1.Close()
	defer client2.Close()
	defer relay1.Close()
	defer relay2.Close()

	var err error

	// Initialize DTLS endpoints
	client1.dtls, err = dtls.NewDTLSClient(keyFile, certFile)
	assert.NotError(t, err, "Error creating client1")

	client2.dtls, err = dtls.NewDTLSClient(keyFile, certFile)
	assert.NotError(t, err, "Error creating client2")

	relay1.dtls, err = dtls.NewDTLSServer(keyFile, certFile)
	assert.NotError(t, err, "Error creating relay1")

	relay2.dtls, err = dtls.NewDTLSServer(keyFile, certFile)
	assert.NotError(t, err, "Error creating relay2")

	// Perform DTLS handshake
	Handshake(t, client1.dtls, relay1.dtls)
	Handshake(t, client2.dtls, relay2.dtls)

	// Set up SRTP based on DTLS
	client1.SetupSRTP(t, false)
	relay1.SetupSRTP(t, true)
	relay2.SetupSRTP(t, true)
	client2.SetupSRTP(t, false)

	// Pass a packet client1 -> relay -> client2
	packetA0, err := srtp.TestPacket(ssrc1, testPacketLen)
	assert.NotError(t, err, "Error creating test packet (A)")

	packetA1, err := client1.srtpOut.Protect(packetA0)
	assert.NotError(t, err, "Error protecting test packet (A 0->1)")

	packetA2, err := relay1.srtpIn.Unprotect(packetA1)
	assert.NotError(t, err, "Error unprotecting test packet (A 1->2)")

	assert.BytesEqual(t, packetA0, packetA2, "C1 -> R1 failed")

	packetA3, err := relay2.srtpOut.Protect(packetA2)
	assert.NotError(t, err, "Error protecting test packet (A 2->3)")

	packetA4, err := client2.srtpIn.Unprotect(packetA3)
	assert.NotError(t, err, "Error unprotecting test packet (A 3->4)")

	assert.BytesEqual(t, packetA2, packetA4, "R2 -> C2 failed")
	assert.BytesEqual(t, packetA0, packetA4, "C1 -> C2 failed")

	// Pass a packet client2 -> relay -> client1
	packetB0, err := srtp.TestPacket(ssrc2, testPacketLen)
	assert.NotError(t, err, "Error creating test packet (B)")

	packetB1, err := client2.srtpOut.Protect(packetB0)
	assert.NotError(t, err, "Error protecting test packet (B 0->1)")

	packetB2, err := relay2.srtpIn.Unprotect(packetB1)
	assert.NotError(t, err, "Error unprotecting test packet (B 1->2)")

	assert.BytesEqual(t, packetB0, packetB2, "C2 -> R2 failed")

	packetB3, err := relay1.srtpOut.Protect(packetB2)
	assert.NotError(t, err, "Error protecting test packet (B 2->3)")

	packetB4, err := client1.srtpIn.Unprotect(packetB3)
	assert.NotError(t, err, "Error unprotecting test packet (B 3->4)")

	assert.BytesEqual(t, packetB2, packetB4, "R1 -> C1 failed")
	assert.BytesEqual(t, packetB0, packetB4, "C2 -> C1 failed")
}
