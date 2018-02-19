package srtp

import (
	"crypto/rand"
	"testing"

	"github.com/bifurcation/percy/assert"
)

func TestSRTPRoundTrip(t *testing.T) {
	profile := AES128CMWith80BitTag
	ssrc := uint32(0x01020304)
	testPacketLen := 30

	keyLen, err := KeyLength(profile)
	assert.NotError(t, err, "Error getting key length")

	saltLen, err := SaltLength(profile)
	assert.NotError(t, err, "Error getting salt length")

	key := make([]byte, keyLen+saltLen)
	rand.Read(key)

	send, err := NewSRTP(AnySSRCOutbound, profile, key)
	assert.NotError(t, err, "Error creating sender")
	defer send.Close()

	recv, err := NewSRTP(AnySSRCInbound, profile, key)
	assert.NotError(t, err, "Error creating receiver")
	defer recv.Close()

	packet0, err := TestPacket(ssrc, testPacketLen)
	assert.NotError(t, err, "Error creating test packet")

	packet1, err := send.Protect(packet0)
	assert.NotError(t, err, "Error encrypting packet")

	packet2, err := recv.Unprotect(packet1)
	assert.NotError(t, err, "Error decrypting packet")

	assert.BytesNotEqual(t, packet0, packet1, "Encryption did not change packet")
	assert.BytesNotEqual(t, packet1, packet2, "Decryption did not change packet")
	assert.BytesEqual(t, packet0, packet2, "Round-trip failed")
}
