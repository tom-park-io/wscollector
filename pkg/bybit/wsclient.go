package bybit

import (
	"time"

	"wscollector/config"
	"wscollector/internal/bybit/memorystore"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WSClient handles WebSocket connection to Bybit and message routing.
type WSClient struct {
	url     string
	args    []string
	conn    *websocket.Conn
	handler func([]byte)
	logger  *zap.Logger
}

// NewClient creates a new WebSocket client with the given URL and logger.
func NewWSClient(url string, logger *zap.Logger) *WSClient {
	return &WSClient{
		url:    url,
		logger: logger,
	}
}

// SetMessageHandler sets the function to handle incoming messages.
func (c *WSClient) SetMessageHandler(h func([]byte)) {
	c.handler = h
}

// Connect establishes the WebSocket connection and subscribes to kline channels
// for all symbols in the provided symbolStore. It does not start the listener.
func (c *WSClient) Connect(cfg *config.Config, symbolStore *memorystore.MemorySymbolStore,
	args []string) error {

	// Attempt to connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		c.logger.Error("Failed to connect to WebSocket", zap.String("url", c.url), zap.Error(err))
		return err
	}
	c.conn = conn
	c.logger.Info("WebSocket connected", zap.String("url", c.url))

	// Store subscription arguments for future reconnects
	c.args = args

	// Send subscription message
	subMsg := map[string]interface{}{
		"op":   "subscribe",
		"args": args,
	}

	if err := conn.WriteJSON(subMsg); err != nil {
		c.logger.Error("Failed to send subscription", zap.Error(err))
		return err
	}

	return nil
}

func (c *WSClient) Listen() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			c.logger.Error("WebSocket read error", zap.Error(err))

			// Retry reconnecting indefinitely
			for {
				time.Sleep(3 * time.Second)
				if err := c.reconnectAndResubscribe(); err != nil {
					c.logger.Warn("Retrying reconnect...")
					continue
				}
				c.logger.Info("Reconnected successfully")
				break
			}
			continue // Start listening again with the new connection
		}

		if c.handler != nil {
			// c.logger.Debug("message received", zap.Int("bytes", len(msg)))
			c.handler(msg)
		}
	}
}

func (c *WSClient) reconnectAndResubscribe() error {
	// Attempt to connect to the WebSocket server
	newConn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}

	// Close the old connection if it exists
	if c.conn != nil {
		_ = c.conn.Close()
	}

	c.conn = newConn

	// Resend the subscription message
	subMsg := map[string]interface{}{
		"op":   "subscribe",
		"args": c.args, // c.args must be stored beforehand
	}
	if err := c.conn.WriteJSON(subMsg); err != nil {
		return err
	}
	return nil
}
