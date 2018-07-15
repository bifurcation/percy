percy
=====

This repo is for a Go module implementing a minimal subset of the
[PERC](https://tools.ietf.org/wg/perc) architecture for secure conferencing.

The overall design has a few pieces:

* An updated version of Firefox that supports EKT and double GCM
* An NSS-based DTLS server (KD) that can negotiate double and EKT
* A simple MD that:
  * Routes DTLS packets between the client and the KD
  * Re-encrypts packets from one client to the others


## Quickstart

```
# Fetch percy through go
> go get github.com/bifurcation/percy
> cd $GOROOT/src/github.com/bifurcation/percy/cmd
> go run main.go

# Pull the appropriate branch of NSS
# Note: This assumes a working build of NSS, see:
#       https://wiki.mozilla.org/NSS/Build_System
> cd $NSS_ROOT
> git remote add ekr https://github.com/ekr/nss.git
> git fetch ekr
> git checkout -b perc_dtls_server ekr/perc_dtls_server
> make nss_build_all
> tests/ssl_gtests/ssl_gtests.sh
> DYLD_LIBRARY_PATH=../dist/$PLATFORM/lib/ ../dist/$PLATFORM/bin/perc_server

# Pull the appropriate branch of Firefox
# Note: This assumes a working build of Firefox, see:
#       https://developer.mozilla.org/en-US/docs/Mozilla/Developer_guide/Build_Instructions
> cd $FIREFOX_ROOT
> git remote add bifurcation https://github.com/bifurcation/gecko-dev
> git checkout -b libsrtp-ekt bifurcation/libsrtp-ekt
> ./mach build
> ./mach run

# Run the example WebRTC app
> cd cmd && go run main.go
# Open in Firefox: https://localhost:4430/
# Click through certificate warning
# Click "Run"
```

