#OPENSSL_INCLUDE=/usr/local/Cellar/openssl/1.0.2n/include/
#OPENSSL_LIB=/usr/local/Cellar/openssl/1.0.2n/lib/

OPENSSL_INCLUDE=${HOME}/Projects/openssl/include
OPENSSL_LIB=${HOME}/Projects/openssl/lib

all:
	gcc -I.. -I${OPENSSL_INCLUDE} -L${OPENSSL_LIB} -lssl -lcrypto ../dtls.c main.c -o dtls

test: all
	./dtls
