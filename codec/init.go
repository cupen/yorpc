package codec

type Codec interface {
	ReadPacket() (bool, uint8, uint16, []byte, error)
	WritePacket(bool, uint8, uint16, []byte) error
}
