#include "x25519_crypto.h"
#include <sodium.h>
#include <string.h>

_Static_assert(crypto_scalarmult_BYTES == KEY_SIZE, "must be 32 bytes");
_Static_assert(crypto_scalarmult_SCALARBYTES == KEY_SIZE, "must be 32 bytes");
_Static_assert(crypto_aead_xchacha20poly1305_ietf_KEYBYTES == KEY_SIZE, "must be 32 bytes");

const char* ErrCodeToStr(int errorCode)
{
	if (errorCode >= ERR_MAX_ID)
	{
		errorCode = 0;
	}

	static const char* errorMessages[ERR_MAX_ID] = {
		"UNKNOWN ERROR",
		"OK",												// ERR_CODE_OK
		"Unable to convert ED25519 public key to X25519",	// ERR_CODE_UNABLE_TO_CONVERT_PUBLIC_KEY
		"Unable to convert ED25519 private key to X25519",	// ERR_CODE_UNABLE_TO_CONVERT_PRIVATE_KEY
		"Unable to make shared secret",						// ERR_CODE_UNABLE_TO_MAKE_SHARED_SECRET
		"Unable to make key from shared secret",			// ERR_CODE_UNABLE_TO_GENERATE_KEY_FROM_SHARED_SECRET
		"Wrong encrypted data size. Use EncryptDataSize()",	// ERR_CODE_WRONG_ENCRYPTED_DATA_SIZE
		"Unable to encrypt data",							// ERR_CODE_UNABLE_TO_ENCRYPT_DATA
		"Wrong decrypted data size. Use DecryptDataSize()",	// ERR_CODE_WRONG_DECRYPTED_DATA_SIZE
		"Unable to decrypt data"							// ERR_CODE_UNABLE_TO_DECRYPT_DATA
	};

	return errorMessages[errorCode];
}

int Ed25519ToX25519
(
	const uint8_t	(*edPublicKey)[KEY_SIZE],
	const uint8_t	(*edPrivateKey)[KEY_SIZE],
	uint8_t			(*xPublicKeyOut)[KEY_SIZE],
	uint8_t			(*xPrivateKeyOut)[KEY_SIZE]
)
{
	// ED25519 public key to X25519
	const int res1 = crypto_sign_ed25519_pk_to_curve25519
	(
		(unsigned char*)xPublicKeyOut,
		(const unsigned char*)edPublicKey
	);

	if (res1 != 0)
	{
		return ERR_CODE_UNABLE_TO_CONVERT_PUBLIC_KEY;
	}

	// ED25519 private key to X25519
	const int res2 = crypto_sign_ed25519_sk_to_curve25519
	(
		(unsigned char*)xPrivateKeyOut,
		(const unsigned char*)edPrivateKey
	);

	if (res2 != 0)
	{
		return ERR_CODE_UNABLE_TO_CONVERT_PRIVATE_KEY;
	}

	return ERR_CODE_OK;
}

int X25519ToSharedSecret
(
	const uint8_t	(*xPublicKeyA)[KEY_SIZE],
	const uint8_t	(*xPrivateKeyB)[KEY_SIZE],
	uint8_t			(*sharedSecretOut)[KEY_SIZE]
)
{
	const int res = crypto_scalarmult
	(
		(unsigned char*)sharedSecretOut,
		(const unsigned char*)xPrivateKeyB,
		(const unsigned char*)xPublicKeyA
	);

	if (res != 0)
	{
		return ERR_CODE_UNABLE_TO_MAKE_SHARED_SECRET;
	}

	return ERR_CODE_OK;
}

int EncryptDataSize (int dataToEncryptSize)
{
	return dataToEncryptSize
		+ crypto_aead_xchacha20poly1305_ietf_ABYTES		// xchacha20poly1305 header
		+ crypto_aead_xchacha20poly1305_ietf_NPUBBYTES;	// nonce
}

int Encrypt
(
	const uint8_t	(*sharedSecret)[KEY_SIZE],
	const uint8_t*	dataToEncrypt,
	int				dataToEncryptSize,
	uint8_t*		encryptedDataOut,
	int				encryptedDataSize
)
{
	// Check encrypted data size
	const int expectedEncryptedDataSize = EncryptDataSize(dataToEncryptSize);
	if (encryptedDataSize != expectedEncryptedDataSize)
	{
		return ERR_CODE_WRONG_ENCRYPTED_DATA_SIZE;
	}

	// Generate key from shared secret
	unsigned char sharedKey[KEY_SIZE];
	{
		const int res = crypto_generichash
		(
			sharedKey,
			KEY_SIZE,
			(const unsigned char*)sharedSecret,
			KEY_SIZE,
			NULL,
			0
		);

		if (res != 0)
		{
			return ERR_CODE_UNABLE_TO_GENERATE_KEY_FROM_SHARED_SECRET;
		}
	}

	// Encrypt message
	{
		unsigned char nonce[crypto_aead_xchacha20poly1305_ietf_NPUBBYTES];
		randombytes_buf(nonce, sizeof(nonce));

		unsigned long long ciphertextLen = 0;
		int res = crypto_aead_xchacha20poly1305_ietf_encrypt
		(
			encryptedDataOut,
			&ciphertextLen,
			(const unsigned char *)dataToEncrypt,
			dataToEncryptSize,
			NULL,
			0, // no additional data
			NULL,
			nonce,
			sharedKey
		);

		if (res != 0)
		{
			return ERR_CODE_UNABLE_TO_ENCRYPT_DATA;
		}

		// Copy nonce
		const int offset = dataToEncryptSize + crypto_aead_xchacha20poly1305_ietf_ABYTES;
		memcpy(encryptedDataOut + offset, nonce, crypto_aead_xchacha20poly1305_ietf_NPUBBYTES);
	}

	return ERR_CODE_OK;
}

int DecryptDataSize (int dataToDecryptSize)
{
	return dataToDecryptSize
		- crypto_aead_xchacha20poly1305_ietf_ABYTES		// xchacha20poly1305 header
		- crypto_aead_xchacha20poly1305_ietf_NPUBBYTES;	// nonce
}

int Decrypt
(
	const uint8_t	(*sharedSecret)[KEY_SIZE],
	const uint8_t*	dataToDecrypt,
	int				dataToDecryptSize,
	uint8_t*		decryptedDataOut,
	int				decryptedDataSize
)
{
	// Check decrypted data size
	const int expectedDecryptedDataSize = DecryptDataSize(dataToDecryptSize);
	if (decryptedDataSize != expectedDecryptedDataSize)
	{
		return ERR_CODE_WRONG_DECRYPTED_DATA_SIZE;
	}

	// Generate key from shared secret
	unsigned char sharedKey[KEY_SIZE];
	{
		const int res = crypto_generichash
		(
			sharedKey,
			KEY_SIZE,
			(const unsigned char*)sharedSecret,
			KEY_SIZE,
			NULL,
			0
		);

		if (res != 0)
		{
			return ERR_CODE_UNABLE_TO_GENERATE_KEY_FROM_SHARED_SECRET;
		}
	}

	// Get nonce
	const int				offset	= decryptedDataSize + crypto_aead_xchacha20poly1305_ietf_ABYTES;
	const unsigned char*	nonce	= ((const unsigned char*)dataToDecrypt) + offset;

	// Decrypt message
	{
		unsigned long long decryptedLen = 0;
		const int res = crypto_aead_xchacha20poly1305_ietf_decrypt
		(
			decryptedDataOut,
			&decryptedLen,
			NULL,
			dataToDecrypt,
			dataToDecryptSize - crypto_aead_xchacha20poly1305_ietf_NPUBBYTES,
			NULL,
			0,
			nonce,
			sharedKey
		);

		if (res != 0)
		{
			return ERR_CODE_UNABLE_TO_DECRYPT_DATA;
		}
	}

	return ERR_CODE_OK;
}
