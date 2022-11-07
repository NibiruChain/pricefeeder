package events

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"sync/atomic"
)

// conn represents a websocket connection interface,
// exists for testing.
type conn interface {
	ReadMessage() (messageType int, msg []byte, err error)
	Close() error
}

type notConnectedWs struct{}

func (notConnectedWs) ReadMessage() (_ int, _ []byte, _ error) {
	return 0, nil, fmt.Errorf("not yet connected")
}
func (notConnectedWs) Close() error { return nil }

func newWs(dialFn func() (conn, error), log zerolog.Logger) *ws {
	ws := &ws{
		log:  log.With().Str("component", "websocket").Logger(),
		stop: make(chan struct{}),
		done: make(chan struct{}),
		read: make(chan []byte),
		dial: dialFn,
		ws:   notConnectedWs{},
	}
	go ws.loop()
	return ws
}

type ws struct {
	log zerolog.Logger

	stop, done chan struct{}

	read chan []byte

	dial func() (conn, error)
	ws   conn
}

func (w *ws) loop() {
	defer close(w.done)

	exit := new(atomic.Bool)
	readLoopDone := make(chan struct{})
	// read votePeriodLoop, reads messages and also handles reconnection
	// alongside first connection too.
	go func() {
		defer close(readLoopDone)
		for {
			_, bytes, err := w.ws.ReadMessage()
			if err != nil {
				// error can be caused by Close
				// so in case the ws was closed we exit
				if exit.Load() {
					return
					// otherwise it's a read error, so we attempt to reconnect. LFG.
				} else {
					w.log.Err(err).Msg("disconnected")
					// we don't care if it fails, because if it does on ReadMessage we will receive an error
					// and then attempt to reconnect again.
					w.connect() // racey with the stop
				}
				// no error, forward the msg
			} else {
				select {
				case w.read <- bytes:
				case <-w.stop:
					w.log.Warn().Str("message", string(bytes)).Msg("message dropped due to shutdown")
					return
				}
			}
		}
	}()

	// wait close votePeriodLoop
	select {
	case <-w.stop:
		exit.Store(true)
		err := w.ws.Close() // this is racey with connect
		if err != nil {
			w.log.Err(err).Msg("close error")
		}
	}
	// wait read votePeriodLoop finished
	<-readLoopDone
}

func (w *ws) connect() {
	var err error
	w.log.Debug().Msg("connecting")
	w.ws, err = w.dial()
	if err != nil {
		w.log.Err(err).Msg("failed to connect")
	} else {
		w.log.Info().Msg("connected")
	}
}

func (w *ws) message() <-chan []byte {
	return w.read
}

func (w *ws) close() {
	close(w.stop)
	<-w.done
}

func dial(url string, onOpenMsg []byte, log zerolog.Logger) *ws {
	return newWs(func() (conn, error) {
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			return nil, err
		}
		err = c.WriteMessage(websocket.BinaryMessage, onOpenMsg)
		return c, err
	}, log)
}
