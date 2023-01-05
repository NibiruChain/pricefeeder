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

func newWebsocket(dialFn dianFn, logger zerolog.Logger) *ws {
	ws := &ws{
		logger:           logger.With().Str("component", "websocket").Logger(),
		stopSignal:       make(chan struct{}),
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
	logger           zerolog.Logger
	stopSignal       chan struct{} // external signal to stop the ws
	done             chan struct{} // internal signal to wait for the ws to execute its shutdown operations
	read             chan []byte
	dial             dianFn
	connection       *websocket.Conn
	connectionClosed *atomic.Bool
}

func (w *ws) loop() {
	defer close(w.done)

	if w.connection == nil {
		w.connect()
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
			w.logger.Err(err).Msg("disconnected from websocket, attempting to reconnect")
			w.connect()
			continue
		}

		// no error, forward the msg
		select {
		case w.read <- bytes:
			w.logger.Debug().Str("payload", string(bytes)).Msg("message received")
		case <-w.stopSignal:
			w.logger.Warn().Str("message", string(bytes)).Msg("message dropped due to shutdown")
			return
		}
	}
}

func (w *ws) connect() {
	w.logger.Debug().Msg("connecting")

	retries := 0
	delay := 1 * time.Second
	for {
		connection, err := w.dial()
		if err == nil {
			w.connection = connection
			w.logger.Debug().Msg("connected to websocket")
			return
		}
		// if we failed to connect, we wait and try again
		w.logger.Err(err).Msg("failed to connect to websocket")
		retries++
		if retries > 10 {
			// if we failed to connect more than 10 times, we exit
			w.logger.Fatal().Msg("failed to connect to websocket")
		}

		w.logger.Debug().Int("retries", retries).Msg("failed to connect to websocket, retrying")
		time.Sleep(delay)
		delay *= 2
	}
}

func (w *ws) message() <-chan []byte {
	return w.read
}

func (w *ws) close() {
	close(w.stopSignal)
	w.connectionClosed.Store(true)
	if err := w.connection.Close(); err != nil {
		w.logger.Err(err).Msg("close error")
	}
	<-w.done
}
