#include "dtls.h"

#include <openssl/err.h>
#include <openssl/ssl.h>
#include <string.h>

// XXX Make these configurable?
#define DTLS_CIPHERS "ALL"
#define SRTP_PROTECTION_PROFILES "SRTP_AES128_CM_SHA1_80"
#define EXPORTER_LABEL "EXTRACTOR-dtls_srtp"
#define EXPORTER_LABEL_LEN strlen(EXPORTER_LABEL)

struct DTLSParamsStr {
  SSL_CTX* ctx;
  SSL* ssl;
  BIO* bio_in;
  BIO* bio_out;
};

int dtls_srtp_ctx(DTLSParams* params, const char* key, const char* cert) {
  int result;

  // Create a new context using DTLS
  params->ctx = SSL_CTX_new(DTLS_method());
  if (params->ctx == NULL) {
    printf("Error: cannot create SSL_CTX.\n");
    ERR_print_errors_fp(stderr);
    return 0;
  }

  // Set our supported ciphers
  result = SSL_CTX_set_cipher_list(params->ctx, DTLS_CIPHERS);
  if (result != 1) {
    printf("Error: cannot set the cipher list.\n");
    ERR_print_errors_fp(stderr);
    return 0;
  }

  // Load the certificate file; contains also the public key
  result = SSL_CTX_use_certificate_file(params->ctx, cert, SSL_FILETYPE_PEM);
  if (result != 1) {
    printf("Error: cannot load certificate file.\n");
    ERR_print_errors_fp(stderr);
    return 0;
  }

  // Load private key
  result = SSL_CTX_use_PrivateKey_file(params->ctx, key, SSL_FILETYPE_PEM);
  if (result != 1) {
    printf("Error: cannot load private key file.\n");
    ERR_print_errors_fp(stderr);
    return 0;
  }

  // Check if the private key is valid
  result = SSL_CTX_check_private_key(params->ctx);
  if (result != 1) {
    printf("Error: checking the private key failed. \n");
    ERR_print_errors_fp(stderr);
    return 0;
  }

  // Turn on SRTP negotiation
  result = SSL_CTX_set_tlsext_use_srtp(params->ctx, SRTP_PROTECTION_PROFILES);
  if (result != 0) {
    printf("Error: Enabling use_srtp failed. \n");
    ERR_print_errors_fp(stderr);
    return 0;
  }

  return 1;
}

int bind(DTLSParams* params) {
  params->bio_in = BIO_new(BIO_s_mem());
  if (params->bio_in == NULL) {
    fprintf(stderr, "error creating input bio\n");
    ERR_print_errors_fp(stderr);
    return 0;
  }

  params->bio_out = BIO_new(BIO_s_mem());
  if (params->bio_out == NULL) {
    fprintf(stderr, "error creating output bio\n");
    ERR_print_errors_fp(stderr);
    return 0;
  }

  params->ssl = SSL_new(params->ctx);
  if (params->ssl == NULL) {
    fprintf(stderr, "error creating SSL\n");
    ERR_print_errors_fp(stderr);
    return 0;
  }

  SSL_set_bio(params->ssl, params->bio_in, params->bio_out);
  return 1;
}

DTLSParams* dtls_client(const char* key, const char* cert) {
  DTLSParams* client = (DTLSParams*)malloc(sizeof(DTLSParams));
  if (!client) {
    return NULL;
  }

  if (!dtls_srtp_ctx(client, key, cert)) {
    return NULL;
  }

  if (!bind(client)) {
    return NULL;
  }

  SSL_set_connect_state(client->ssl);
  return client;
}

DTLSParams* dtls_server(const char* key, const char* cert) {
  DTLSParams* server = (DTLSParams*)malloc(sizeof(DTLSParams));
  if (!server) {
    return NULL;
  }

  if (!dtls_srtp_ctx(server, key, cert)) {
    return NULL;
  }

  if (!bind(server)) {
    return NULL;
  }

  bind(server);
  SSL_set_accept_state(server->ssl);
  return server;
}

void dtls_free(DTLSParams* params) {
  if (params->ctx) {
    SSL_CTX_free(params->ctx);
  }

  if (params->ssl) {
    // This also frees the underlying BIOs
    SSL_free(params->ssl);
  }

  free(params);
}

int dtls_kick(DTLSParams* params) {
  int ret = SSL_do_handshake(params->ssl);
  if (ret < 0) {
    int err = SSL_get_error(params->ssl, ret);
    ERR_print_errors_fp(stderr);
  }
  return ret;
}

void dtls_send(DTLSParams* params, void* packet, size_t packet_size) {
  BIO_write(params->bio_in, packet, packet_size);
}

int dtls_recv(DTLSParams* params, void* packet, size_t max_size) {
  return BIO_read(params->bio_out, packet, max_size);
}

int dtls_done(DTLSParams* params) { return SSL_is_init_finished(params->ssl); }

uint16_t dtls_srtp_profile(DTLSParams* params) {
  SRTP_PROTECTION_PROFILE* profile = SSL_get_selected_srtp_profile(params->ssl);
  return profile->id;
}

void dtls_srtp_key(DTLSParams* params, uint8_t* key, size_t key_len) {
  SSL_export_keying_material(params->ssl, key, key_len, EXPORTER_LABEL,
                             EXPORTER_LABEL_LEN, NULL, 0, 0);
}

