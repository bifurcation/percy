#ifndef __DTLS_H__
#define __DTLS_H__

#include <stdlib.h>

typedef struct DTLSParamsStr DTLSParams;

DTLSParams* dtls_client(const char* key, const char* cert);
DTLSParams* dtls_server(const char* key, const char* cert);
void dtls_free(DTLSParams* params);
void dtls_kick(DTLSParams* params);
void dtls_send(DTLSParams* params, void* packet, size_t packet_size);
int dtls_recv(DTLSParams* params, void* packet, size_t max_size);
int dtls_done(DTLSParams* params);
uint16_t dtls_srtp_profile(DTLSParams* params);
void dtls_srtp_key(DTLSParams* params, uint8_t* key, size_t key_len);

#endif // ndef __DTLS_H__
