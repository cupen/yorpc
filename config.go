package yorpc

import (
	"fmt"
	"time"
)

type Options struct {
	MaxConn        int // 最大连接数
	HeartBeat      int // 心跳间隔(秒)
	ReadBufferSize int // 缓存的消息数量
}

func (opts *Options) Check() {
	if opts.MaxConn <= 0 {
		panic(fmt.Errorf("invalid MaxConn:%d", opts.MaxConn))
	}

	if opts.HeartBeat < 60 {
		panic(fmt.Errorf("invalid HeartBeat:%d", opts.HeartBeat))
	}
}

func (opts *Options) GetHeartBeatDrt() time.Duration {
	if opts == nil {
		return 0
	}
	return time.Duration(opts.HeartBeat) * time.Second
}
