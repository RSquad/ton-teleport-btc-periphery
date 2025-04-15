#include <iostream>
#include <iomanip>
#include <vector>
#include <array>
#include <sodium.h>
#include "../x25519_crypto.h"

template<typename T>
std::string to_hex_str(const T& aData)
{
	std::ostringstream s;
	s << std::hex
	  << std::setw(2)
	  << std::setfill('0');

	for (const auto& byte: aData)
	{
		s << static_cast<int>(byte);
	}

	return s.str();
}

using EdPublicKeyT	= std::array<unsigned char, crypto_sign_PUBLICKEYBYTES>;
using EdPrivateKeyT	= std::array<unsigned char, crypto_sign_SECRETKEYBYTES>;
using XPublicKeyT	= std::array<unsigned char, crypto_scalarmult_BYTES>;
using XPrivateKeyT	= std::array<unsigned char, crypto_scalarmult_SCALARBYTES>;

int main()
{
	if (sodium_init() < 0)
	{
		std::cerr << "libsodium initialization failed!" << std::endl;
		return EXIT_FAILURE;
	}

	// Generate Ed25519 keys for Alice and Bob
	EdPublicKeyT	alicePublicKey_Ed25519;
	EdPrivateKeyT	alicePrivateKey_Ed25519;
	if (crypto_sign_keypair(alicePublicKey_Ed25519.data(), alicePrivateKey_Ed25519.data()) != 0)
	{
		std::cerr << "crypto_sign_keypair failed" << std::endl;
		return EXIT_FAILURE;
	}

	EdPublicKeyT	bobPublicKey_Ed25519;
	EdPrivateKeyT	bobPrivateKey_Ed25519;
	if (crypto_sign_keypair(bobPublicKey_Ed25519.data(), bobPrivateKey_Ed25519.data()) != 0)
	{
		std::cerr << "crypto_sign_keypair failed" << std::endl;
		return EXIT_FAILURE;
	}

	// Convert Ed25519 keys to X25519 (Alice)
	XPublicKeyT		alicePublicKey_X25519;
	XPrivateKeyT	alicePrivateKey_X25519;
	{
		const int code = Ed25519ToX25519
		(
			(const uint8_t(*)[32])(alicePublicKey_Ed25519.data()),
			(const uint8_t(*)[32])(alicePrivateKey_Ed25519.data()),
			(uint8_t(*)[32])(alicePublicKey_X25519.data()),
			(uint8_t(*)[32])(alicePrivateKey_X25519.data())
		);

		if (code != ERR_CODE_OK)
		{
			std::cerr << ErrCodeToStr(code) << std::endl;
			return EXIT_FAILURE;
		}
	}

	// Convert Ed25519 keys to X25519 (Bob)
	XPublicKeyT		bobPublicKey_X25519;
	XPrivateKeyT	bobPrivateKey_X25519;
	{
		const int code = Ed25519ToX25519
		(
			(const uint8_t(*)[32])(bobPublicKey_Ed25519.data()),
			(const uint8_t(*)[32])(bobPrivateKey_Ed25519.data()),
			(uint8_t(*)[32])(bobPublicKey_X25519.data()),
			(uint8_t(*)[32])(bobPrivateKey_X25519.data())
		);

		if (code != ERR_CODE_OK)
		{
			std::cerr << ErrCodeToStr(code) << std::endl;
			return EXIT_FAILURE;
		}
	}

	// Alice shared secret
	XPublicKeyT aliceSharedSecret;
	{
		const int code = X25519ToSharedSecret
		(
			(const uint8_t(*)[32])(bobPublicKey_X25519.data()),
			(const uint8_t(*)[32])(alicePrivateKey_X25519.data()),
			(uint8_t(*)[32])(aliceSharedSecret.data())
		);

		if (code != ERR_CODE_OK)
		{
			std::cerr << ErrCodeToStr(code) << std::endl;
			return EXIT_FAILURE;
		}
	}

	// Bob shared secret
	XPublicKeyT bobSharedSecret;
	{
		const int code = X25519ToSharedSecret
		(
			(const uint8_t(*)[32])(alicePublicKey_X25519.data()),
			(const uint8_t(*)[32])(bobPrivateKey_X25519.data()),
			(uint8_t(*)[32])(bobSharedSecret.data())
		);

		if (code != ERR_CODE_OK)
		{
			std::cerr << ErrCodeToStr(code) << std::endl;
			return EXIT_FAILURE;
		}
	}

	// Verify shared secrets are the same
	std::cout << "Alice's Shared Secret: '" << to_hex_str(aliceSharedSecret) << "'" << std::endl;
	std::cout << "Bob's Shared Secret:   '" << to_hex_str(bobSharedSecret) << "'" << std::endl;

	if (aliceSharedSecret != bobSharedSecret)
	{
		std::cerr << "Alice's shared secret does not match Bob's" << std::endl;
		return EXIT_FAILURE;
	}

	// Encrypt data (by Alice)
	const std::string			plainMessage{"Hello Bob!"};
	const int					encryptedDataSize = EncryptDataSize(plainMessage.size());
	std::vector<unsigned char>	encryptedMessage(encryptedDataSize);
	{
		const int code = Encrypt
		(
			(const uint8_t(*)[32])(aliceSharedSecret.data()),
			(const uint8_t*)plainMessage.data(),
			(int)plainMessage.size(),
			(uint8_t*)encryptedMessage.data(),
			(int)encryptedMessage.size()
		);

		if (code != ERR_CODE_OK)
		{
			std::cerr << ErrCodeToStr(code) << std::endl;
			return EXIT_FAILURE;
		}
	}

	std::cout << "Encrypted Message: '" << to_hex_str(encryptedMessage) << "'" << std::endl;

	// Decrypt message (Bob side)
	const int			decryptedDataSize = DecryptDataSize(encryptedMessage.size());
	std::vector<char>	decryptedMessage(decryptedDataSize);
	{
		const int code = Decrypt
		(
			(const uint8_t(*)[32])(bobSharedSecret.data()),
			(const uint8_t*)encryptedMessage.data(),
			(int)encryptedMessage.size(),
			(uint8_t*)decryptedMessage.data(),
			(int)decryptedMessage.size()
		);

		if (code != ERR_CODE_OK)
		{
			std::cerr << ErrCodeToStr(code) << std::endl;
			return EXIT_FAILURE;
		}
	}

	std::string_view decryptedPlainMessage{decryptedMessage.data(), decryptedMessage.size()};
	std::cout << "Plain Message:     '" << decryptedPlainMessage << "'" << std::endl;
	std::cout << "Decrypted Message: '" << decryptedPlainMessage << "'" << std::endl;

	if (decryptedPlainMessage != plainMessage)
	{
		std::cerr << "Decrypted message must be equal to plain message" << std::endl;
		return EXIT_FAILURE;
	}

	return EXIT_SUCCESS;
}
