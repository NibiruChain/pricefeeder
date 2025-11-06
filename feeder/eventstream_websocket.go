package feeder

import (
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type dianWsFn func() (*websocket.Conn, error)

// ws represents a WebSocket connection with automatic reconnection capabilities.
// It handles connection establishment, message reading, and graceful shutdown.
type ws struct {
	logger           zerolog.Logger
	stopSignal       chan struct{} // external signal to stop the ws
	done             chan struct{} // internal signal to wait for the ws to execute its shutdown operations
	read             chan []byte
	dial             dianWsFn
	connection       *websocket.Conn
	connectionClosed *atomic.Bool
}

// NewWebsocket creates a new WebSocket connection to the specified URL.
// The connection automatically sends onOpenMsg as a binary message immediately after
// establishing the connection. The function returns a WebSocket instance that runs
// in a background goroutine, handling connection, reconnection, and message reading.
//
// Parameters:
//   - url: The WebSocket server URL (e.g., "wss://echo.websocket.org")
//   - onOpenMsg: A binary message to send immediately after connection is established
//   - logger: A zerolog logger instance for logging connection events and errors
//
// The WebSocket will automatically attempt to reconnect with exponential backoff
// if the connection is lost (up to 10 retry attempts). Messages can be read from
// the channel returned by the message() method.
func NewWebsocket(url string, onOpenMsg []byte, logger zerolog.Logger) *ws {
	dialFunction := func() (*websocket.Conn, error) {
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			return nil, err
		}
		return conn, conn.WriteMessage(websocket.BinaryMessage, onOpenMsg)
	}

	ws := &ws{
		logger:           logger.With().Str("component", "websocket").Logger(),
		stopSignal:       make(chan struct{}),
		done:             make(chan struct{}),
		read:             make(chan []byte),
		dial:             dialFunction,
		connection:       nil,
		connectionClosed: new(atomic.Bool),
	}

	go ws.loop()
	return ws
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
			w.logger.Warn().Str("payload", string(bytes)).Msg("message dropped due to shutdown")
			return
		}
	}
}

// connect attempts to dial the websocket using binary exponential backoff.
// It will retry up to 10 times with exponentially increasing delays (1s, 2s, 4s, etc.)
// before giving up. On successful connection, it sends the onOpenMsg and begins reading messages.
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

// Message returns a read-only channel that receives all messages from the WebSocket connection.
// Messages are delivered as []byte. The channel will be closed when the WebSocket connection
// is closed via the Close() method.
func (w *ws) Message() <-chan []byte {
	return w.read
}

// Close gracefully closes the WebSocket connection. It signals the connection loop to stop,
// closes the underlying WebSocket connection, and waits for all shutdown operations to complete.
// This method is safe to call multiple times and will not panic.
func (w *ws) Close() {
	close(w.stopSignal)
	w.connectionClosed.Store(true)
	if w.connection != nil {
		if err := w.connection.Close(); err != nil {
			w.logger.Err(err).Msg("close error")
		}
	}
	<-w.done
}
