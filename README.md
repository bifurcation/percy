percy
=====

This repo is for a Go module implementing a minimal subset of the
[PERC](https://tools.ietf.org/wg/perc) architecture for secure conferencing.

Right now, we have the following components available:

* An skeleton MD in Go that can discriminate between STUN, DTLS, and
  SRTP packets and forward them appropriately.  In the long run:
  * STUN packets should be handled directly by the MD
  * DTLS packets should be forwarded between the client and the KD
  * SRTP packets should be broadcast to conference participants

* A CGo wrapper around [OpenSSL](https://github.com/openssl/openssl)
  that provides DTLS negotiation.

* A CGo wrapper around [libsrtp](https://github.com/cisco/libstrp)
  that provides SRTP encryption / decryption.

* A small WebRTC app that demonstrates one-to-one media between two
  peers, relayed via this server.

Right now, 1-1 "conferencing" works, via the simple WebRTC app
included.  But that's just because the only thing the server does
right now is switch packets.  (It doesn't even give different
treatment to different packet classes.)

In order to get conferencing working (even without PERC), we would
need:

* STUN termination, which requires reading the SDP offer sent by a
  connecting endpoint toget the ICE `ufrag` and `pwd` attributes

* DTLS termination, which requires synthesizing the SDP answer for
  the conference to set the `setup` and `fingerprint` attributes.

* SRTP re-encryption, to bridge between different DTLS associations.

Once we have those pieces in place, transitioning to PERC is a
simple matter of the KD lying to the MD -- telling it to use AES-GCM
when in fact AES-GCM-double was negotiated.


## Quickstart

```
# Fetch and build C dependencies
> git submodule update --init
> cd third-party/openssl && ./config -static && make && cd ../..
> cd third-party/libsrtp && ./configure && make && cd ../..

# Build and run self-tests
> go build ./...
> go test ./...

# Run the example WebRTC app
> cd cmd && go run main.go
# Open in Firefox: https://localhost:4430/
# Click through certificate warning
# Click "Run"
# If you get two videos, it worked
```

