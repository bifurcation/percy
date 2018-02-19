#include "go_srtp.h"

#include <stdio.h>
#include <string.h>

int main() {
  if (go_srtp_init() != 0) {
    fprintf(stderr, "Error initializing libsrtp\n");
    return 1;
  }

  int ciphersuite = srtp_profile_aes128_cm_sha1_80;
  size_t key_len = 30;
  uint8_t key[30] = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x10, 0x11,
                     0x12, 0x13, 0x14, 0x15, 0x20, 0x21, 0x22, 0x23,
                     0x24, 0x25, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35};

  srtp_t send = go_srtp_create(ssrc_any_outbound, ciphersuite, key, key_len);
  if (!send) {
    fprintf(stderr, "Error initializing sender\n");
    return 1;
  }

  srtp_t recv = go_srtp_create(ssrc_any_inbound, ciphersuite, key, key_len);
  if (!recv) {
    fprintf(stderr, "Error initializing receiver\n");
    return 1;
  }

  uint8_t packet_send[1024], packet_ct[1024], packet_recv[1024];
  size_t packet_len_send = 30, packet_len_ct, packet_len_recv;

  if (!go_srtp_test_packet(0x12345678, packet_send, packet_len_send)) {
    fprintf(stderr, "Error creating test packet\n");
    return 1;
  }

  memcpy(packet_ct, packet_send, packet_len_send);
  packet_len_ct = go_srtp_protect(send, packet_ct, packet_len_send);
  if (packet_len_ct == 0) {
    fprintf(stderr, "Error in go_srtp_protect\n");
    return 1;
  }

  memcpy(packet_recv, packet_ct, packet_len_ct);
  packet_len_recv = go_srtp_unprotect(recv, packet_recv, packet_len_ct);
  if (packet_len_recv == 0) {
    fprintf(stderr, "Error in go_srtp_unprotect\n");
    return 1;
  }

  if (packet_len_send != packet_len_recv) {
    fprintf(stderr, "Received packet has wrong length %lu != %lu\n",
            packet_len_send, packet_len_recv);
    return 1;
  }

  if (memcmp(packet_send, packet_recv, packet_len_recv) != 0) {
    fprintf(stderr, "Sent and received packets are different\n");
    return 1;
  }

  printf("PASS\n");
  return 0;
}
