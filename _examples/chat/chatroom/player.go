package chatroom

type PlayerImpl struct {
	id string
}

func NewPlayerImpl(id string) Player {
	obj := PlayerImpl{
		id: id,
	}
	return &obj
}

func (p *PlayerImpl) Speak(args []byte) ([]byte, error) {
	// p.records = append(p.records, string(args))
	return args, nil
}

func (p *PlayerImpl) SpeakAsync(args []byte) error {
	// p.records = append(p.records, string(args))
	return nil
}
