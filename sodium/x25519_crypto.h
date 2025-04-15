#pragma once

#include <stdint.h>

// Error codes
#define ERR_CODE_OK											1
#define ERR_CODE_UNABLE_TO_CONVERT_PUBLIC_KEY				2
#define ERR_CODE_UNABLE_TO_CONVERT_PRIVATE_KEY				3
#define ERR_CODE_UNABLE_TO_MAKE_SHARED_SECRET				4
#define ERR_CODE_UNABLE_TO_GENERATE_KEY_FROM_SHARED_SECRET	5
#define ERR_CODE_WRONG_ENCRYPTED_DATA_SIZE					6
#define ERR_CODE_UNABLE_TO_ENCRYPT_DATA						7
#define ERR_CODE_WRONG_DECRYPTED_DATA_SIZE					8
#define ERR_CODE_UNABLE_TO_DECRYPT_DATA						9
#define ERR_MAX_ID											10

#define KEY_SIZE 32

#ifdef __cplusplus
extern "C" {
#endif

const char* ErrCodeToStr(int errorCode);

int Ed25519ToX25519
(
	const uint8_t	(*edPublicKey)[KEY_SIZE],
	const uint8_t	(*edPrivateKey)[KEY_SIZE],
	uint8_t			(*xPublicKeyOut)[KEY_SIZE],
	uint8_t			(*xPrivateKeyOut)[KEY_SIZE]
);

int X25519ToSharedSecret
(
	const uint8_t	(*xPublicKeyA)[KEY_SIZE],
	const uint8_t	(*xPrivateKeyB)[KEY_SIZE],
	uint8_t			(*sharedSecretOut)[KEY_SIZE]
);

int EncryptDataSize (int dataToEncryptSize);

int Encrypt
(
	const uint8_t	(*sharedSecret)[KEY_SIZE],
	const uint8_t*	dataToEncrypt,
	int				dataToEncryptSize,
	uint8_t*		encryptedDataOut,
	int				encryptedDataSize
);

int DecryptDataSize (int dataToDecryptSize);

int Decrypt
(
	const uint8_t	(*sharedSecret)[KEY_SIZE],
	const uint8_t*	dataToDecrypt,
	int				dataToDecryptSize,
	uint8_t*		decryptedDataOut,
	int				decryptedDataSize
);

#ifdef __cplusplus
} // extern "C"
#endif
