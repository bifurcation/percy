percy
=====

This repo is for a Go module implementing chunks of the
[PERC](https://tools.ietf.org/wg/perc) architecture for secure conferencing.
Things that are in progress now:

* An MDD that can selectively forward DTLS packets to a KMF and switch SRTP
  packets between endpoints.

Things that might be done in the future:

* A KMF that can set up associations with endpoints

* An endpoint implementation

## Quickstart

```
# Fetch and build C dependencies
> git submodule update --init
> cd third-party/openssl && ./config && make && cd ../..
> cd third-party/libsrtp && ./configure && make && cd ../..

# Build and run self-tests
> go build ./...
> go test ./...
```

