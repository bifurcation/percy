#ifndef __SRTP_H__
#define __SRTP_H__

#include <srtp.h>
#include <stdlib.h>

// Some convenient values defined in srtp.h:
//
// Error codes:
// srtp_err_status_ok - success error code
//
// Protection profiles:
// srtp_profile_aes128_cm_sha1_80
// srtp_profile_aes128_cm_sha1_32
//
// SSRC types:
// ssrc_any_inbound
// ssrc_any_outbound

int go_srtp_init();
int go_srtp_key_length(int profile);
int go_srtp_salt_length(int profile);
int go_srtp_create(srtp_ctx_t** session, int type, int profile,
                   const uint8_t* key, size_t len);
void go_srtp_free(srtp_ctx_t* session);
int go_srtp_protect(srtp_ctx_t* session, uint8_t* packet, int len);
int go_srtp_unprotect(srtp_ctx_t* session, uint8_t* packet, int len);

int go_srtp_test_packet(uint32_t ssrc, uint8_t* packet, int len);

#endif  // ndef __SRTP_H__
