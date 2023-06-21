package codecv1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeCall(t *testing.T) {
	assert := assert.New(t)
	codec := CodecV1{}

	msg := codec.EncodeCall(1, 2, []byte{4, 5, 6})
	assert.Equal([]byte{0b10000001, 2, 0, 4, 5, 6}, msg)

	msg = codec.EncodeCall(127, 0xFA01, []byte{7, 8, 9})
	assert.Equal([]byte{0b11111111, 0x1, 0xFA, 7, 8, 9}, msg)
}

func BenchmarkEncodeCall(b *testing.B) {
	codec := CodecV1{}
	for i := 0; i < b.N; i++ {
		msg := codec.EncodeCall(127, 0xFA01, []byte{7, 8, 9})
		if len(msg) != 6 {
			b.FailNow()
		}
	}
}
