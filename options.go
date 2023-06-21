package yorpc

import (
	"fmt"
	"time"
)

func DefaultOptions() *Options {
	return &Options{
		DebugLog:       false,
		MaxConn:        100_000,
		HeartBeat:      10,
		ReadBufferSize: 16,
	}
}

type Options struct {
	DebugLog       bool
	MaxConn        int
	HeartBeat      int // seconds
	ReadBufferSize int // 缓存的消息数量
	OnError        func(uint16, []byte, error)
}

func (this *Options) Check() {
	if this.MaxConn <= 0 {
		panic(fmt.Errorf("invalid MaxConn:%d", this.MaxConn))
	}

	if this.HeartBeat <= 0 {
		panic(fmt.Errorf("invalid HeartBeat:%d", this.HeartBeat))
	}
}

func (this *Options) GetHeartBeatDrt() time.Duration {
	if this == nil {
		return 0
	}
	return time.Duration(this.HeartBeat) * time.Second
}
