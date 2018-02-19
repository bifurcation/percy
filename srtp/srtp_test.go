package srtp

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"
)

func assert(t *testing.T, test bool, message string) {
	if !test {
		t.Fatal(message)
	}
}

func assertNotError(t *testing.T, err error, msg string) {
	assert(t, err == nil, fmt.Sprintf("%s: %v", msg, err))
}

func assertBytesEqual(t *testing.T, a, b []byte, msg string) {
	assert(t, bytes.Equal(a, b), fmt.Sprintf("%s: [%x] != [%x]", msg, a, b))
}

func assertBytesNotEqual(t *testing.T, a, b []byte, msg string) {
	assert(t, !bytes.Equal(a, b), fmt.Sprintf("%s: [%x] =q= [%x]", msg, a, b))
}

func TestSRTPRoundTrip(t *testing.T) {
	profile := AES128CMWith80BitTag
	ssrc := uint32(0x01020304)
	testPacketLen := 30

	keyLen, err := KeyLength(profile)
	assertNotError(t, err, "Error getting key length")

	key := make([]byte, keyLen)
	rand.Read(key)

	send, err := NewSRTP(AnySSRCOutbound, profile, key)
	assertNotError(t, err, "Error creating sender")
	defer send.Close()

	recv, err := NewSRTP(AnySSRCInbound, profile, key)
	assertNotError(t, err, "Error creating receiver")
	defer recv.Close()

	packet0, err := TestPacket(ssrc, testPacketLen)
	assertNotError(t, err, "Error creating test packet")

	packet1, err := send.Protect(packet0)
	assertNotError(t, err, "Error encrypting packet")

	packet2, err := recv.Unprotect(packet1)
	assertNotError(t, err, "Error decrypting packet")

	assertBytesNotEqual(t, packet0, packet1, "Encryption did not change packet")
	assertBytesNotEqual(t, packet1, packet2, "Decryption did not change packet")
	assertBytesEqual(t, packet0, packet2, "Round-trip failed")
}
