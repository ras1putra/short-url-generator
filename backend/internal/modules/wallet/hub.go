package wallet

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"urlshortener/pkg/logger"
)

type Hub struct {
	clients    map[uuid.UUID]map[*websocket.Conn]bool
	register   chan *client
	unregister chan *client
	broadcast  chan *Notification
	mu         sync.RWMutex
}

type client struct {
	userID uuid.UUID
	conn   *websocket.Conn
}

type Notification struct {
	UserID  uuid.UUID   `json:"-"`
	Type    string      `json:"type"` // e.g. "WALLET_UPDATE"
	Payload interface{} `json:"payload,omitempty"`
}

var GlobalHub = &Hub{
	clients:    make(map[uuid.UUID]map[*websocket.Conn]bool),
	register:   make(chan *client),
	unregister: make(chan *client),
	broadcast:  make(chan *Notification),
}

func (h *Hub) Start(ctx context.Context) {
	logger.Ctx(ctx).Info("WebSocket Wallet Notification Hub started")

	for {
		select {
		case <-ctx.Done():
			logger.Ctx(ctx).Info("WebSocket Wallet Notification Hub stopped")
			return
		case c := <-h.register:
			h.mu.Lock()
			if h.clients[c.userID] == nil {
				h.clients[c.userID] = make(map[*websocket.Conn]bool)
			}
			h.clients[c.userID][c.conn] = true
			h.mu.Unlock()
			logger.Ctx(ctx).Debug("WS client registered", zap.String("user_id", c.userID.String()))

		case c := <-h.unregister:
			h.mu.Lock()
			if h.clients[c.userID] != nil {
				delete(h.clients[c.userID], c.conn)
				if len(h.clients[c.userID]) == 0 {
					delete(h.clients, c.userID)
				}
			}
			h.mu.Unlock()
			logger.Ctx(ctx).Debug("WS client unregistered", zap.String("user_id", c.userID.String()))

		case n := <-h.broadcast:
			h.mu.RLock()
			conns, ok := h.clients[n.UserID]
			if ok {
				msgBytes, err := json.Marshal(n)
				if err == nil {
					for conn := range conns {
						err := conn.WriteMessage(websocket.TextMessage, msgBytes)
						if err != nil {
							logger.Ctx(ctx).Warn("Failed to send WS message to client, unregistering",
								zap.String("user_id", n.UserID.String()),
								zap.Error(err),
							)
							_ = conn.Close()
							go func(connToClose *websocket.Conn) {
								h.unregister <- &client{userID: n.UserID, conn: connToClose}
							}(conn)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToUser(userID uuid.UUID, nType string, payload interface{}) {
	// Safely pass to channel without blocking
	select {
	case h.broadcast <- &Notification{
		UserID:  userID,
		Type:    nType,
		Payload: payload,
	}:
	default:
		// Drop notification if channel buffer is full
	}
}
