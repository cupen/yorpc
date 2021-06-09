package yorpc

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testCase1() (http.HandlerFunc, chan []byte) {
	ch := make(chan []byte)
	f := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("test case request")
		conn, err := NewWebsocketByHTTP(r, w)
		if err != nil {
			panic(err)
		}

		hub := NewHandlersHub()
		hub.RegisterRPC(1, func(args []byte) ([]byte, error) {
			return args, nil
		})

		hub.RegisterRPCAsync(2, func(args []byte) error {
			return nil
		})
		s := NewServer(conn, hub, nil)

		log.Printf("test case running")
		if err := s.Run(); err != nil {
			panic(err)
		}
		log.Printf("test case stopped")
	}
	return f, ch
}

func TestExample(t *testing.T) {
	assert := assert.New(t)
	h, ch := testCase1()
	s := httptest.NewServer(http.HandlerFunc(h))

	// l, err := net.Listen("tcp", "127.0.0.1:0")
	// assert.NoError(err)

	l := s.Listener
	// go s.Start()
	time.Sleep(100 * time.Millisecond)

	t.Cleanup(func() {
		defer s.Close()
		l.Close()
		close(ch)
	})

	url := fmt.Sprintf("ws://%s/", l.Addr().String())
	client, err := NewClient(url)
	client.Start()
	assert.NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := client.Call(ctx, 1, []byte{1, 2, 3})
	assert.NoError(err)
	assert.Equal([]byte{1, 2, 3}, resp)

	// select {
	// case rs := <-ch:
	// 	assert.NotNil(rs)
	// case <-ctx.Done():
	// 	assert.NoError(ctx.Err())
	// 	assert.Fail("timeout")
	// }
}
