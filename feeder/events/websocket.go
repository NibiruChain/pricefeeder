package events

import (
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type dianFn func() (*websocket.Conn, error)

func NewWebsocket(url string, onOpenMsg []byte, logger zerolog.Logger) *ws {
	return newWebsocket(func() (*websocket.Conn, error) {
		connection, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			return nil, err
		}
		return connection, connection.WriteMessage(websocket.BinaryMessage, onOpenMsg)
	}, logger)
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

	if w.connection == nil {
		// if the connection is nil, then we attempt to connect using binary exponential backoff
		attempt := 0
		delay := 1 * time.Second
		for {
			w.connect()
			if w.connection != nil {
				break
			}

			// if we failed to connect, we wait and try again
			attempt++
			if attempt > 10 {
				// if we failed to connect more than 10 times, we exit
				w.log.Fatal().Msg("failed to connect to websocket")
			}

			w.log.Debug().Int("attempt", attempt).Msg("failed to connect to websocket, retrying")
			time.Sleep(delay)
			delay *= 2
		}
	}

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
			w.log.Err(err).Msg("disconnected from websocket, attempting to reconnect")
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
		w.log.Err(err).Msg("failed to connect to websocket")
	} else {
		w.connection = connection
		w.log.Debug().Msg("connected to websocket")
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
