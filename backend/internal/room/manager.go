package room

import (
	"log/slog"
	"sync"

	"lil-poker/internal/types"
)

type WSManager struct {
	clientsMu     sync.RWMutex
	clients       map[*WsClient]bool
	sg            *SafeGame
	StartingChips int
}

func NewWSManager(sg *SafeGame, startingChips int) *WSManager {
	return &WSManager{
		clients:       make(map[*WsClient]bool),
		sg:            sg,
		StartingChips: startingChips,
	}
}

func (wsm *WSManager) Register(client *WsClient) {
	wsm.clientsMu.Lock()
	defer wsm.clientsMu.Unlock()
	wsm.clients[client] = true
}

func (wsm *WSManager) Unregister(client *WsClient) {
	wsm.clientsMu.Lock()
	deleted := false
	if _, ok := wsm.clients[client]; ok {
		delete(wsm.clients, client)
		deleted = true
	}
	wsm.clientsMu.Unlock()

	if deleted {
		client.Close()
	}
}

func (wsm *WSManager) GetObserverCount() int {
	wsm.clientsMu.RLock()
	defer wsm.clientsMu.RUnlock()

	wsm.sg.mu.RLock()
	players := wsm.sg.game.Players
	playerIDs := make(map[string]bool)
	for _, p := range players {
		playerIDs[p.ID] = true
	}
	wsm.sg.mu.RUnlock()

	count := 0
	for client := range wsm.clients {
		if !playerIDs[client.PlayerID] {
			count++
		}
	}
	return count
}

func (wsm *WSManager) Broadcast() {
	snap := wsm.sg.getSnapshot()
	observerCount := wsm.GetObserverCount()
	observers := wsm.GetObserverNames()

	wsm.clientsMu.RLock()
	defer wsm.clientsMu.RUnlock()

	for client := range wsm.clients {
		resp := snap.toResponse(client.PlayerID, observerCount, wsm.StartingChips)
		resp.Observers = observers
		select {
		case client.Send <- resp:
		default:
			slog.Warn("WS send queue full, dropping update")
		}
	}
}

func (wsm *WSManager) GetObserverNames() []string {
	wsm.clientsMu.RLock()
	defer wsm.clientsMu.RUnlock()

	wsm.sg.mu.RLock()
	players := wsm.sg.game.Players
	playerIDs := make(map[string]bool)
	for _, p := range players {
		playerIDs[p.ID] = true
	}
	wsm.sg.mu.RUnlock()

	nameMap := make(map[string]bool)
	for client := range wsm.clients {
		if !playerIDs[client.PlayerID] && client.Username != "" {
			nameMap[client.Username] = true
		}
	}

	names := make([]string, 0, len(nameMap))
	for name := range nameMap {
		names = append(names, name)
	}
	return names
}

func (wsm *WSManager) DisconnectPlayer(playerID string) {
	if playerID == "" {
		return
	}
	wsm.clientsMu.Lock()
	var clientsToDisconnect []*WsClient
	for client := range wsm.clients {
		if client.PlayerID == playerID {
			clientsToDisconnect = append(clientsToDisconnect, client)
			delete(wsm.clients, client)
		}
	}
	wsm.clientsMu.Unlock()

	for _, client := range clientsToDisconnect {
		collisionMsg := types.GameStateResponse{
			Phase: "Collision",
		}
		select {
		case client.Send <- collisionMsg:
		default:
		}
		client.Close()
	}
}

func (wsm *WSManager) IsPlayerConnected(playerID string) bool {
	wsm.clientsMu.RLock()
	defer wsm.clientsMu.RUnlock()
	for client := range wsm.clients {
		if client.PlayerID == playerID {
			return true
		}
	}
	return false
}

func (wsm *WSManager) Close() {
	closedState := types.GameStateResponse{
		Phase: "Closed",
	}

	wsm.clientsMu.Lock()
	clientsCopy := make([]*WsClient, 0, len(wsm.clients))
	for client := range wsm.clients {
		select {
		case client.Send <- closedState:
		default:
		}
		clientsCopy = append(clientsCopy, client)
		delete(wsm.clients, client)
	}
	wsm.clientsMu.Unlock()

	for _, client := range clientsCopy {
		client.Close()
	}
}
