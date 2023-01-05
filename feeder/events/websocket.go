package events

import (
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type dianFn func() (*websocket.Conn, error)

func NewWebsocket(url string, onOpenMsg []byte, log zerolog.Logger) *ws {
	return newWebsocket(func() (*websocket.Conn, error) {
		connection, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			return nil, err
		}
		return connection, connection.WriteMessage(websocket.BinaryMessage, onOpenMsg)
	}, log)
}

func newWebsocket(dialFn dianFn, log zerolog.Logger) *ws {
	ws := &ws{
		log:              log.With().Str("sub-component", "websocket").Logger(),
		done:             make(chan struct{}),
		read:             make(chan []byte),
		dial:             dialFn,
		connection:       nil,
		connectionClosed: new(atomic.Bool),
	}

	go ws.loop()
	return ws
}

type ws struct {
	log              zerolog.Logger
	done             chan struct{} // internal signal to wait for the ws to execute its shutdown operations
	read             chan []byte
	dial             dianFn
	connection       *websocket.Conn
	connectionClosed *atomic.Bool
}

func (w *ws) loop() {
	defer close(w.done)

	w.connect()

	// read messages and also handles reconnection.
	for {
		_, bytes, err := w.connection.ReadMessage()
		if err != nil {
			if w.connectionClosed.Load() {
				// if the connection was closed, then we exit
				return
			}

			// otherwise we attempt to reconnect
			// we don't care if it fails, because if it does on ReadMessage we will receive an error
			// and then attempt to reconnect again.
			w.log.Err(err).Msg("disconnected")
			w.connect()
			continue
		}

		// no error, forward the msg
		w.read <- bytes
		w.log.Debug().Str("payload", string(bytes)).Msg("message received")
	}
}

func (w *ws) connect() {
	w.log.Debug().Msg("connecting")
	connection, err := w.dial()
	if err != nil {
		w.log.Err(err).Msg("failed to connect")
	} else {
		w.connection = connection
		w.log.Debug().Msg("connected")
	}
}

func (w *ws) message() <-chan []byte {
	return w.read
}

func (w *ws) close() {
	w.connectionClosed.Store(true)
	if err := w.connection.Close(); err != nil {
		w.log.Err(err).Msg("close error")
	}
	<-w.done
}
