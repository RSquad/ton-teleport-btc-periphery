package utils

func SliceOfBytesContains(haystack [][]byte, needle []byte) bool {
	bytesMap := make(map[string]bool, len(haystack))
	for _, bytes := range haystack {
		key := string(bytes)
		bytesMap[key] = true
	}
	_, exists := bytesMap[string(needle)]
	return exists
}
