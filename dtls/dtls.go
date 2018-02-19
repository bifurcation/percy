package dtls

/*
#cgo darwin CFLAGS: -I../third-party/openssl/include
#cgo darwin LDFLAGS: -L../third-party/openssl/ -lssl -lcrypto

#include "dtls.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

const (
	mtu = 1024
)

type DTLS struct {
	params *C.struct_DTLSParamsStr
}

func NewDTLSClient(key, cert string) (*DTLS, error) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	cCert := C.CString(cert)
	defer C.free(unsafe.Pointer(cCert))

	params := C.dtls_client(cKey, cCert)
	if params == nil {
		return nil, errors.New("Could not allocate DTLS params")
	}
	return &DTLS{params}, nil
}

func NewDTLSServer(key, cert string) (*DTLS, error) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	cCert := C.CString(cert)
	defer C.free(unsafe.Pointer(cCert))

	params := C.dtls_server(cKey, cCert)
	if params == nil {
		return nil, errors.New("Could not allocate DTLS params")
	}
	return &DTLS{params}, nil
}

func (d *DTLS) Close() {
	if d != nil && d.params != nil {
		C.dtls_free(d.params)
		d.params = nil
	}
}

func (d *DTLS) Kick() {
	C.dtls_kick(d.params)
}

func (d *DTLS) Send(packet []byte) {
	C.dtls_send(d.params, unsafe.Pointer(&packet[0]), C.size_t(len(packet)))
}

func (d *DTLS) Recv() []byte {
	packet := make([]byte, mtu)
	packetLen := int(C.dtls_recv(d.params, unsafe.Pointer(&packet[0]), mtu))

	if packetLen < 0 {
		packetLen = 0
	}

	return packet[:int(packetLen)]
}

func (d *DTLS) Done() bool {
	return (C.dtls_done(d.params) == 1)
}

func (d *DTLS) SRTPProfile() uint16 {
	return uint16(C.dtls_srtp_profile(d.params))
}

func (d *DTLS) SRTPKey(size int) []byte {
	key := make([]byte, size)
	ptr := (*C.uint8_t)(unsafe.Pointer(&key[0]))
	C.dtls_srtp_key(d.params, ptr, C.size_t(size))
	return key
}
