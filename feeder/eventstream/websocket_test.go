package eventstream

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestWebsocketSuccess(t *testing.T) {
	ws := NewWebsocket("wss://echo.websocket.events/.ws", []byte("test"), zerolog.New(os.Stderr))
	defer ws.close()
	// LOL this test websocket URL we're using returns the following
	select {
	case msg := <-ws.message():
		require.Equal(t, []byte("echo.websocket.events sponsored by Lob.com"), msg)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
	// test send as we will receive an echo
	select {
	case msg := <-ws.message():
		require.Equalf(t, []byte("test"), msg, string(msg))
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func TestWebsocketExplicitClose(t *testing.T) {
	ws := NewWebsocket("wss://echo.websocket.events/.ws", []byte("test"), zerolog.New(os.Stderr))
	require.NotPanics(t, func() {
		ws.close()
	})
}
