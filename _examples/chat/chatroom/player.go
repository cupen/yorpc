package chatroom

import "log"

var _ ChatroomAPI = &PlayerImpl{}

type PlayerImpl struct {
	id string
}

func NewPlayerImpl(id string) *PlayerImpl {
	obj := PlayerImpl{
		id: id,
	}
	return &obj
}

func (p *PlayerImpl) Speak(msg *Message) (*void, error) {
	log.Printf("speak: %#v", msg)
	return &void{}, nil
}

func (p *PlayerImpl) SpeakAsync(msg *Message) (*void, error) {
	log.Printf("speak async: %#v", msg)
	return &void{}, nil
}
