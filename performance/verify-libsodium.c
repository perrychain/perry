
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <time.h>

#include "common.h"
#include "sodium.h"

#define MAX_MSG_LEN 1024

// Verify a pre-defined secret/public key and signature for a sample implementation
int main()
{
    /*
    // Max size for a Perry payload
    uint8_t m[376];

    unsigned char *public_key;
    unsigned char *signature;
    // unsigned char *message;

    const char *base64_1, *base64_2, *base64_3;

    // Specify the message
    // !! For correct signature calculation, must include \0 and the entire length to compare
    // First, set the mem to \0 for the message
    memset(m, '\0', sizeof m);

    base64_1 = "MjAyMi0wNy0wNSAyMDo1MDowNy43NzA4NjggKzAyMDAgQ0VTVCBtPSswLjAxNDYwMDgzNCBXT1cgQSBzdXBlciBpbXBvcnRhbnQgbWVzc2FnZSBoZXJlIGZvciB5b3UgdG8gY29uc2lkZXIgMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==";

    // message = mybase64_decode(base64_1);
    int mlen = snprintf((char *)m, 376, "%s", mybase64_decode(base64_1));

    printf("\nMessage (len=%d): %s\n", mlen, m);

    // Sender Public key
    base64_2 = "gv1qBsdMCOqOuUCgTFodqsjmOEgVhPzKiDpaeRplyUg=";
    public_key = mybase64_decode(base64_2);

    // Signature
    base64_3 = "piWhlx6C3jlnSfEnQUPUVSAlzbKWjL0mtAhvg+YUQITjGGku1aW/9R+FqAzXztH4vJwfyz0vURkRvlZRDLNpAA==";
    signature = mybase64_decode(base64_3);

    int status;

    // Verify the provided signature matches the public key
    status = crypto_sign_ed25519_verify_detached((unsigned char *)signature, m, sizeof(m), public_key);

    printf("\nVerify valid signature status => %d\n", status);

    if (status == 0)
    {
        printf("\tVerify OK (Pass)\n");
    }
    else
    {
        printf("\tVerify error\n");
    }

    // Verify a fake signature does not pass verification
    status = crypto_sign_verify_detached((unsigned char *)"XYZ", m, strlen((char *)m), public_key);

    printf("\nVerifiy invalid signature status => %d\n", status);

    if (status == 0)
    {
        printf("\tVerify OK (error!)\n");
    }
    else
    {
        printf("\tVerify did not pass (OK, expected)\n");
    }
    */

    // Benchmarking tests
    static unsigned char sk[crypto_box_SECRETKEYBYTES];
    static unsigned char pk[crypto_box_PUBLICKEYBYTES];

#define MESSAGE (const unsigned char *) "Hello, world!"
#define MESSAGE_LEN 4

    unsigned char sig[crypto_sign_BYTES];

    clock_t start;
    clock_t end;
    int i;

    printf("testing seed generation performance: ");
    unsigned char seed[32];
    start = clock();
    for (i = 0; i < 10000; ++i) {
        randombytes_buf(seed, sizeof seed);
    }
    end = clock();

    printf("%fus per seed\n", ((double) ((end - start) * 1000)) / CLOCKS_PER_SEC / i * 1000);

    printf("testing key generation performance: ");
    start = clock();
    for (i = 0; i < 10000; ++i)
    {
        if( crypto_box_keypair(pk, sk) != 0) {
            printf("Key generation failed!\n");
        }
    }
    end = clock();

    printf("%fus per keypair\n", ((double)((end - start) * 1000)) / CLOCKS_PER_SEC / i * 1000);


    printf("testing sign performance: ");
    start = clock();
    for (i = 0; i < 10000; ++i) {
        crypto_sign_keypair(pk, sk);
        if(crypto_sign_detached(sig, NULL, MESSAGE, MESSAGE_LEN, sk) != 0) {
            printf("Signing failed!\n");
        }
    }
    end = clock();

    printf("%fus per signature\n", ((double) ((end - start) * 1000)) / CLOCKS_PER_SEC / i * 1000);

    printf("testing verify performance: ");
    start = clock();
    for (i = 0; i < 10000; ++i) {
        if( crypto_sign_verify_detached(sig, MESSAGE, MESSAGE_LEN, pk) != 0) {
            printf("Verification failed!\n");
        }
    }
    end = clock();

    printf("%fus per signature\n", ((double) ((end - start) * 1000)) / CLOCKS_PER_SEC / i * 1000);
    

    return 0;

}
