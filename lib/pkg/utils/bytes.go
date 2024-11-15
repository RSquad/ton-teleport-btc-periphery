package utils

func BytesPadTo(input []byte, size int) []byte {
	padded := make([]byte, size)
	copy(padded[size-len(input):], input)
	return padded
}
