package room

import (
	"sync"

	"github.com/gorilla/websocket"

	"lil-poker/internal/types"
)

type WsClient struct {
	Conn      *websocket.Conn
	PlayerID  string
	Username  string
	Send      chan types.GameStateResponse
	closeOnce sync.Once
}

func NewWsClient(conn *websocket.Conn, playerID, username string) *WsClient {
	return &WsClient{
		Conn:     conn,
		PlayerID: playerID,
		Username: username,
		Send:     make(chan types.GameStateResponse, 64),
	}
}

func (c *WsClient) Close() {
	c.closeOnce.Do(func() {
		close(c.Send)
		_ = c.Conn.Close()
	})
}
