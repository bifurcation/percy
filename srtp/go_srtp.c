#include "go_srtp.h"
#include <arpa/inet.h>
#include <srtp.h>
#include <stdio.h>
#include <string.h>

#define SRTP_AES128_CM_SHA1_80 0x0001
#define SRTP_AES128_CM_SHA1_32 0x0002

const size_t RTP_HEADER_SIZE = 12;
const size_t SRTP_TRAILER_SIZE = SRTP_MAX_TRAILER_LEN;

int go_srtp_init() { return srtp_init(); }

int go_srtp_key_length(int profile) {
  return srtp_profile_get_master_key_length(profile);
}

int go_srtp_salt_length(int profile) {
  return srtp_profile_get_master_salt_length(profile);
}

srtp_t go_srtp_create(int type, int profile, const uint8_t* key, size_t len) {
  int err;

  srtp_policy_t policy;
  memset(&policy, 0, sizeof(policy));

  err = srtp_crypto_policy_set_from_profile_for_rtp(&policy.rtp, profile);
  if (err != srtp_err_status_ok) {
    fprintf(stderr, "Unable to set crypto policy for RTP %d\n", err);
    return NULL;
  }

  err = srtp_crypto_policy_set_from_profile_for_rtcp(&policy.rtcp, profile);
  if (err != srtp_err_status_ok) {
    fprintf(stderr, "Unable to set crypto policy for RTCP %d\n", err);
    return NULL;
  }

  // Cargo-culted from webrtc.org: srtpfilter.cc
  policy.ssrc.type = (srtp_ssrc_type_t)(type);
  policy.ssrc.value = 0;
  policy.key = (uint8_t*)(key);
  policy.window_size = 1024;
  policy.allow_repeat_tx = 1;
  policy.next = NULL;

  srtp_t session = NULL;
  err = srtp_create(&session, &policy);
  if (err != srtp_err_status_ok) {
    fprintf(stderr, "Failed to create SRTP session %d\n", err);
    return NULL;
  }

  return session;
}

void go_srtp_free(srtp_t session) { srtp_dealloc(session); }

// Caller should make sure that packet has SRTP_MAX_TRAILER_LEN
// octets free after the end of the RTP data, into which
// srtp_protect can write.
int go_srtp_protect(srtp_t session, uint8_t* packet, int len) {
  int out_len = len;
  int err = srtp_protect(session, packet, &out_len);
  if (err != srtp_err_status_ok) {
    fprintf(stderr, "Error encrypting SRTP packet %d\n", err);
    return 0;
  }

  return out_len;
}

int go_srtp_unprotect(srtp_t session, uint8_t* packet, int len) {
  int out_len = len;
  int err = srtp_unprotect(session, packet, &out_len);
  if (err != srtp_err_status_ok) {
    fprintf(stderr, "Error decrypting SRTP packet %d\n", err);
    return 0;
  }

  return out_len;
}

int go_srtp_test_packet(uint32_t ssrc, uint8_t* packet, int len) {
  if (len < RTP_HEADER_SIZE) {
    fprintf(stderr, "Buffer too small for an RTP packet\n");
    return 0;
  }

  packet[0] = 0x80;                              // v=2 p=x=cc=0
  packet[1] = 0x7f;                              // m=0 pt=0x7t
  *(uint16_t*)(packet + 2) = htons(0x1234);      // seq
  *(uint32_t*)(packet + 4) = htonl(0xdecafbad);  // ts
  *(uint32_t*)(packet + 8) = htonl(ssrc);

  for (int i = RTP_HEADER_SIZE; i < len; ++i) {
    packet[i] = 0xab;
  }
  return len;
}
