package codec

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXOR(t *testing.T) {
	expected := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i := 1; i <= len(data); i++ {
		t.Run(fmt.Sprintf("XOR-%d", i), func(t *testing.T) {
			assert := require.New(t)
			key := make([]byte, i)
			_, err := rand.Read(key)
			assert.NoError(err)
			assert.Equal(expected, data)
			xor(data, key)
			assert.NotEqual(expected, data)
			xor(data, key)
			assert.Equal(expected, data)
		})
	}
}
