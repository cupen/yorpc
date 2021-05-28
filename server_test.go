package yorpc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/stretchr/testify/assert"
)

func testCase1(w http.ResponseWriter, r *http.Request) {
	s := NewServer(nil)
	if err := s.Start(r, w); err != nil {
		panic(err)
	}
}

func TestExample(t *testing.T) {
	assert := assert.New(t)

	s := httptest.NewServer(http.HandlerFunc(testCase1))
	t.Cleanup(func() {
		defer s.Close()
	})

	ctx := context.Background()
	conn, _, _, err := ws.Dial(ctx, "ws://127.0.0.1/")
	assert.NoError(err)
	err = wsutil.WriteClientBinary(conn, nil)
	assert.NoError(err)
}
