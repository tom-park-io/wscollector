package bybit

import (
	"fmt"
	"time"

	"wscollector/internal/bybit/memorystore"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WSClient handles WebSocket connection to Bybit and message routing.
type WSClient struct {
	url         string
	args        []string
	conn        *websocket.Conn
	handler     func([]byte)
	symbolStore *memorystore.MemorySymbolStore
	logger      *zap.Logger
}

// NewClient creates a new WebSocket client with the given URL and logger.
func NewWSClient(url string, store *memorystore.MemorySymbolStore, logger *zap.Logger) *WSClient {
	return &WSClient{
		url:         url,
		symbolStore: store,
		logger:      logger,
	}
}

// SetMessageHandler sets the function to handle incoming messages.
func (c *WSClient) SetMessageHandler(h func([]byte)) {
	c.handler = h
}

// Connect establishes the WebSocket connection and subscribes to kline channels
// for all symbols in the provided symbolStore. It does not start the listener.
func (c *WSClient) Connect() error {

	// Attempt to connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		c.logger.Error("Failed to connect to WebSocket", zap.String("url", c.url), zap.Error(err))
		return err
	}
	c.conn = conn
	c.logger.Info("WebSocket connected", zap.String("url", c.url))

	// Store subscription arguments for future reconnects
	c.args = c.symbolStore.GetKlineTopics(c.symbolStore.WsInterval)

	// Send subscription message
	subMsg := map[string]interface{}{
		"op":   "subscribe",
		"args": c.args,
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

	// Replace the current connection
	c.conn = newConn

	// Regenerate subscription topics based on current symbols
	c.args = c.symbolStore.GetKlineTopics(c.symbolStore.WsInterval)

	// Build subscription message payload
	subMsg := map[string]interface{}{
		"op":   "subscribe",
		"args": c.args, // c.args must be stored beforehand
	}

	// Send the subscription message
	if err := c.conn.WriteJSON(subMsg); err != nil {
		return fmt.Errorf("websocket subscribe failed: %w", err)
	}

	return nil
}
