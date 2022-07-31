// libsodium
#include <sodium.h>

typedef struct {
        unsigned char* Parent;
        int64_t SeqID;
        unsigned char* SeqTime;
}Header;

typedef struct {
        const char* Sender;
        const char* Recipient;
        const char* Signature;
        const char* Data;
        const char* Header;
        int Type;
        int Reserved;
        unsigned char* Output;
        int64_t Block;
}TxPayload;

typedef struct {
    Header header;
    TxPayload payload[];

}Block;

typedef struct {
    unsigned char* key;
    Block value;

}BlockKV;


// Common functions

int sign(uint8_t sm[], const uint8_t m[], const int mlen, const uint8_t sk[]) {
	unsigned long long smlen;

	if( crypto_sign_ed25519(sm,&smlen, m, mlen, sk) == 0) {
		return smlen;
	} else {
		return -1;
	}
}

int verify(uint8_t m[], const uint8_t sm[], const int smlen, const uint8_t pk[]) {
	unsigned long long mlen;

	if( crypto_sign_open(m, &mlen, sm, smlen, pk) == 0) {
		return mlen;
	} else {
		return -1;
	}
}

char* to_hex( char hex[], const uint8_t bin[], size_t length )
{
	int i;
	uint8_t *p0 = (uint8_t *)bin;
	char *p1 = hex;

	for( i = 0; i < length; i++ ) {
		snprintf( p1, 3, "%02x", *p0 );
		p0 += 1;
		p1 += 2;
	}

	return hex;
}


unsigned char * mybase64_decode(const char *base64)
{
    const char *base64_end;
    size_t bin_len;

    char unsigned *data = calloc(512, sizeof(data));

    if (sodium_base642bin(data, 512 * sizeof(data), base64, strlen(base64), "", &bin_len, &base64_end, sodium_base64_VARIANT_ORIGINAL_NO_PADDING) == -1)
    {
        printf("sodium_base642bin failed\n");
    }

    return data;
}
