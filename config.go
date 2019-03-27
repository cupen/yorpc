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

func (this *Options) Check() {
	if this.MaxConn <= 0 {
		panic(fmt.Errorf("Invalid MaxConn:%d", this.MaxConn))
	}

	if this.HeartBeat < 60 {
		panic(fmt.Errorf("Invalid HeartBeat:%d", this.HeartBeat))
	}
}

func (this *Options) GetHeartBeatDrt() time.Duration {
	if this == nil {
		return 0
	}
	return time.Duration(this.HeartBeat) * time.Second

}
