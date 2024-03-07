package codec

func xor(input, key []byte) {
	var length = min(len(input), len(key))
	for i := 0; i < length; i++ {
		input[i] ^= key[i]
	}
}
