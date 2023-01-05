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
		log:        log.With().Str("sub-component", "websocket").Logger(),
		stopSignal: make(chan struct{}),
		done:       make(chan struct{}),
		read:       make(chan []byte),
		dial:       dialFn,
		connection: nil,
	}
	go ws.loop()
	return ws
}

type ws struct {
	log        zerolog.Logger
	stopSignal chan struct{} // allows external callers to stop the ws
	done       chan struct{} // internal signal to wait for the ws to execute its shutdown operations
	read       chan []byte
	dial       dianFn
	connection *websocket.Conn
}

func (w *ws) loop() {
	defer close(w.done)

	w.connect()

	exit := new(atomic.Bool)
	readLoopDone := make(chan struct{})

	// readLoop reads messages and also handles reconnection alongside first connection too.
	go func() {
		defer close(readLoopDone)

		for {
			_, bytes, err := w.connection.ReadMessage()
			if err != nil {
				// error can be caused by Close
				// so in case the ws was closed we exit
				if exit.Load() {
					return
				}

				// otherwise it's a read error, so we attempt to reconnect. LFG.
				w.log.Err(err).Msg("disconnected")
				// we don't care if it fails, because if it does on ReadMessage we will receive an error
				// and then attempt to reconnect again.
				w.connect() // racey with connection Close()
				continue
			}

			// no error, forward the msg
			select {
			case w.read <- bytes:
			case <-w.stopSignal:
				w.log.Warn().Str("message", string(bytes)).Msg("message dropped due to shutdown")
				return
			}
		}
	}()

	// wait for a stop signal
	<-w.stopSignal
	exit.Store(true)
	if err := w.connection.Close(); err != nil { // this is racey with connect
		w.log.Err(err).Msg("close error")
	}

	// wait readLoop finished
	<-readLoopDone
}

func (w *ws) connect() {
	w.log.Debug().Msg("connecting")
	connection, err := w.dial()
	if err != nil {
		w.log.Err(err).Msg("failed to connect")
	} else {
		w.connection = connection
		w.log.Info().Msg("connected")
	}
}

func (w *ws) message() <-chan []byte {
	return w.read
}

func (w *ws) close() {
	close(w.stopSignal)
	<-w.done
}
