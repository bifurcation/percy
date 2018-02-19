package srtp

/*
#cgo darwin CFLAGS: -I../third-party/libsrtp/include
#cgo darwin LDFLAGS: -L../third-party/libsrtp -lsrtp2

#include "go_srtp.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

var (
	libsrtpInitialized = false
)

const (
	maxSRTPTrailer = int(C.SRTP_MAX_TRAILER_LEN)
	srtpErrOK      = 0
)

const (
	AES128CMWith80BitTag = int(C.srtp_profile_aes128_cm_sha1_80)
	AES128CMWith32BitTag = int(C.srtp_profile_aes128_cm_sha1_32)

	AnySSRCInbound  = int(C.ssrc_any_inbound)
	AnySSRCOutbound = int(C.ssrc_any_outbound)
)

func KeyLength(profile int) (int, error) {
	keyLen := int(C.go_srtp_key_length(C.int(profile)))
	if keyLen == 0 {
		return 0, errors.New("Unknown ciphersuite")
	}
	return keyLen, nil
}

func SaltLength(profile int) (int, error) {
	saltLen := int(C.go_srtp_salt_length(C.int(profile)))
	if saltLen == 0 {
		return 0, errors.New("Unknown ciphersuite")
	}
	return saltLen, nil
}

type SRTP struct {
	ctx *C.struct_srtp_ctx_t_
}

func NewSRTP(ssrc_type, profile int, key []byte) (*SRTP, error) {
	if !libsrtpInitialized {
		C.go_srtp_init()
		libsrtpInitialized = true
	}

	s := SRTP{}
	ptr := (*C.uint8_t)(unsafe.Pointer(&key[0]))

	s.ctx = C.go_srtp_create(C.int(ssrc_type), C.int(profile), ptr, C.size_t(len(key)))
	if s.ctx == nil {
		return nil, errors.New("Could not create libsrtp context")
	}
	return &s, nil
}

func (s *SRTP) Close() {
	if s == nil || s.ctx == nil {
		return
	}

	C.go_srtp_free(s.ctx)
}

func (s *SRTP) Protect(packet []byte) ([]byte, error) {
	out := make([]byte, len(packet)+maxSRTPTrailer)
	copy(out, packet)
	ptr := (*C.uint8_t)(unsafe.Pointer(&out[0]))
	out_len := C.go_srtp_protect(s.ctx, ptr, C.int(len(packet)))
	if out_len == 0 {
		return nil, errors.New("Error encrypting SRTP packet")
	}
	return out[:out_len], nil
}

func (s *SRTP) Unprotect(packet []byte) ([]byte, error) {
	out := make([]byte, len(packet))
	copy(out, packet)
	ptr := (*C.uint8_t)(unsafe.Pointer(&out[0]))
	out_len := C.go_srtp_unprotect(s.ctx, ptr, C.int(len(out)))
	if out_len == 0 {
		return nil, errors.New("Error decrypting SRTP packet")
	}
	return out[:out_len], nil
}

func TestPacket(ssrc uint32, size int) ([]byte, error) {
	out := make([]byte, size)
	ptr := (*C.uint8_t)(unsafe.Pointer(&out[0]))
	if C.go_srtp_test_packet(C.uint32_t(ssrc), ptr, C.int(size)) == 0 {
		return nil, errors.New("Error creating test SRTP packet")
	}
	return out, nil
}
