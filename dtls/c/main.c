#include "dtls.h"
#include <stdio.h>

#define CERT_FILE "../cert.pem"
#define KEY_FILE "../key.pem"
#define MAX_PACKET 1024
#define KEY_SIZE 60

int main() {
  DTLSParams* client = dtls_client(KEY_FILE, CERT_FILE);
  DTLSParams* server = dtls_server(KEY_FILE, CERT_FILE);

  uint8_t packet[MAX_PACKET];
  int num_bytes = 0;
  int total_bytes = 0;
  int round = 0;

  do {
    printf("%dxRTT - ", round++);

    dtls_kick(client);

    // Send C->S flight
    total_bytes = 0;
    do {
      num_bytes = dtls_recv(client, packet, MAX_PACKET);
      dtls_send(server, packet, num_bytes);
      total_bytes += num_bytes;
    } while (num_bytes > 0);
    printf("c2s->[%d] ", total_bytes);

    dtls_kick(server);

    // Send S->C flight
    total_bytes = 0;
    do {
      num_bytes = dtls_recv(server, packet, MAX_PACKET);
      dtls_send(client, packet, num_bytes);
      total_bytes += num_bytes;
    } while (num_bytes > 0);
    printf("s2c->[%d] ", total_bytes);

    printf("\n");
    getchar();
  } while (!dtls_done(client) || !dtls_done(server));

  uint8_t client_key[KEY_SIZE], server_key[KEY_SIZE];
  dtls_srtp_key(client, client_key, KEY_SIZE);
  dtls_srtp_key(server, server_key, KEY_SIZE);

  printf("client key: ");
  for (int i = 0; i < KEY_SIZE; ++i) {
    printf("%02x", client_key[i]);
  }
  printf("\n");

  printf("server key: ");
  for (int i = 0; i < KEY_SIZE; ++i) {
    printf("%02x", server_key[i]);
  }
  printf("\n");

  dtls_free(client);
  dtls_free(server);

  return 0;
}
