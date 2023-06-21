package yorpc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/cupen/yorpc/connection/websocket"
	"github.com/cupen/yorpc/handlerhub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type _case interface {
	Logf(string, ...interface{})
	Cleanup(func())
}

func testCase1(t _case, hub *handlerhub.Hub) (http.HandlerFunc, chan []byte) {
	ch := make(chan []byte)
	f := func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.NewServer(r, w)
		if err != nil {
			panic(err)
		}
		s := NewSession(conn, hub, nil)
		t.Logf("testcase1 starting")
		if err := s.Run(); err != nil {
			panic(err)
		}
		t.Logf("testcase1 stopped")
	}
	t.Cleanup(func() {
		close(ch)
	})
	return f, ch
}

func TestCall(t *testing.T) {
	assert := assert.New(t)

	hub := handlerhub.New()
	hub.Register(1, func(args []byte) ([]byte, error) {
		return args, nil
	})
	hub.Register(2, func(args []byte) ([]byte, error) {
		if len(args) <= 0 {
			return nil, nil
		}
		return args[0 : len(args)-1], nil
	})
	hub.Register(3, func(args []byte) ([]byte, error) {
		if len(args) <= 0 {
			return nil, nil
		}
		return args[len(args)-1:], nil
	})

	countor := 0
	hub.Register2(11, func(args []byte) {
		countor++
		return
	})
	h, _ := testCase1(t, hub)
	s := httptest.NewServer(http.HandlerFunc(h))
	t.Cleanup(func() {
		defer s.Close()
	})

	url := fmt.Sprintf("ws://%s/", s.Listener.Addr())
	conn, err := websocket.NewClient(url, 2*time.Second)
	assert.NoError(err)

	client := NewClient(conn, 2*time.Second)
	makeBytes := func() []byte {
		rs := make([]byte, 256)
		for i := 0; i < 256; i++ {
			rs[i] = byte(i)
		}
		return rs
	}

	resp, err := client.Call(nil, 1, []byte{1, 2, 3})
	assert.NoError(err)
	assert.Equal([]byte{1, 2, 3}, resp)

	S := fmt.Sprintf
	data := makeBytes()
	for i := 1; i < 255; i++ {
		args := data[0:i]
		t.Run(S("call-1(size=%d)", len(args)), func(t *testing.T) {
			assert := require.New(t)
			resp, err := client.Call(nil, 1, args)
			assert.NoError(err)
			assert.Equal(args, resp)
		})

		t.Run(S("call-2(size=%d)", len(args)), func(t *testing.T) {
			assert := require.New(t)
			resp, err := client.Call(nil, 2, args)
			assert.NoError(err)
			assert.Equal(args[0:len(args)-1], resp)
		})
		t.Run(S("call-3(size=%d)", len(args)), func(t *testing.T) {
			assert := require.New(t)
			resp, err := client.Call(nil, 3, args)
			assert.NoError(err)
			assert.Equal(args[len(args)-1:], resp)
		})
	}
	for i := 1; i < 255; i++ {
		args := data[0:i]
		t.Run(S("send(size=%d)", len(args)), func(t *testing.T) {
			assert := require.New(t)
			countor = 0
			assert.Equal(0, countor)
			for j := 0; j < 10; j++ {
				client.Send(11, args)
			}
			client.Call(nil, 1, args)
			assert.Equal(10, countor)

			for j := 0; j < 100; j++ {
				client.Send(110, args)
			}
			client.Call(nil, 1, args)
			assert.Equal(10, countor)
		})
	}
}

func BenchmarkClientWithHttptest(b *testing.B) {
	hub := handlerhub.New()
	hub.Register(999, func(args []byte) ([]byte, error) {
		return args, nil
	})
	countor := 0
	hub.Register2(1999, func(args []byte) {
		countor++
		return
	})

	h, _ := testCase1(b, hub)
	s := httptest.NewServer(http.HandlerFunc(h))
	b.Cleanup(func() {
		defer s.Close()
	})

	url := fmt.Sprintf("ws://%s/", s.Listener.Addr())
	conn, err := websocket.NewClient(url, 2*time.Second)
	if err != nil {
		b.Fatalf("new client failed. err:%v", err)
	}
	client := NewClient(conn, 2*time.Second)
	makeBytes := func(size int) []byte {
		rs := make([]byte, size)
		for i := 0; i < size; i++ {
			rs[i] = byte(i % 256)
		}
		return rs
	}
	data := makeBytes(10000)
	sizes := []int{1, 10, 100, 1000, 10000}
	for _, n := range sizes {
		args := data[0:n]
		runtime.GC()
		b.Run(fmt.Sprintf("call(size=%d)", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resp, err := client.Call(nil, 999, args)
				if err != nil {
					b.Fatalf("call failed. err:%v", err)
				}
				if len(resp) != len(args) {
					b.Fatalf("call failed. resp(%d) != args(%d)", len(resp), len(args))
				}
			}
		})
	}

	for _, n := range sizes {
		args := data[0:n]
		runtime.GC()
		b.Run(fmt.Sprintf("send(size=%d)", n), func(b *testing.B) {
			countor = 0
			for i := 0; i < b.N; i++ {
				client.Send(1999, args)
			}
			resp, err := client.Call(nil, 999, args)
			if err != nil || len(resp) != len(args) {
				b.Fatalf("call failed. err:%v", err)
			}
			b.Logf("countor=%d b.N=%d", countor, b.N)
			if countor != b.N {
				b.Fatalf("countor(%d) != b.N(%d)", countor, b.N)
			}
		})
	}
}
