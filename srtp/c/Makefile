LIBSRTP_INCLUDE=${HOME}/Projects/libsrtp/include
LIBSRTP_LIB=${HOME}/Projects/libsrtp/

all:
	gcc -I.. -I${LIBSRTP_INCLUDE} -L${LIBSRTP_LIB} -lsrtp2 ../go_srtp.c main.c -o srtp

test: all
	./srtp
