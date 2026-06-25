package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"lil-poker/internal/game"
	"lil-poker/internal/room"
	"lil-poker/internal/store"
	"lil-poker/internal/types"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

func (s *Server) roomFromRequest(w http.ResponseWriter, r *http.Request) *room.Room {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		writeError(w, http.StatusBadRequest, "missing room query parameter")
		return nil
	}
	r2, ok := s.rm.GetRoom(roomID)
	if !ok {
		writeError(w, http.StatusNotFound, "room not found")
		return nil
	}
	return r2
}

func (s *Server) registerClient(r *room.Room, client *room.WsClient) {
	r.Wsm.Register(client)
}

func (s *Server) unregisterClient(r *room.Room, client *room.WsClient) {
	r.Wsm.Unregister(client)

	if client.PlayerID == r.CreatorID && r.CreatorID != "" {
		go func() {
			time.Sleep(10 * time.Second)
			if !r.Wsm.IsPlayerConnected(r.CreatorID) {
				slog.Info("Creator disconnected, removing room", "room_id", r.ID, "creator_id", r.CreatorID)
				s.rm.DeleteRoom(r.ID)
			}
		}()
	}

	if client.PlayerID != "" {
		go func(pID string) {
			time.Sleep(10 * time.Second)
			if !r.Wsm.IsPlayerConnected(pID) {
				r.Sg.Lock()
				g := r.Sg.GetGame()
				var pName string
				isPlayerActive := false
				for _, p := range g.Players {
					if p.ID == pID {
						if !p.SittingOut {
							pName = p.Name
							if g.Phase != game.PhaseWaiting && g.Phase != game.PhaseShowdown && g.ActiveIdx >= 0 && g.ActiveIdx < len(g.Players) && g.Players[g.ActiveIdx].ID == pID {
								isPlayerActive = true
							}
							_ = g.SitOut(pID)
							r.Sg.UpdateHandEvaluations()
						}
						break
					}
				}
				r.Sg.Unlock()

				if pName != "" {
					slog.Info("Player disconnected, forced sit-out", "room_id", r.ID, "player", pName)
					r.Sg.AddSystemMessage(fmt.Sprintf("%s was sat out due to disconnection.", pName))
					s.broadcast(r)
					if isPlayerActive {
						s.resetTimer(r)
					}
				}
			}
		}(client.PlayerID)
	}
}

func (s *Server) writePump(r *room.Room, client *room.WsClient) {
	ticker := time.NewTicker(pingPeriod)
	var unregOnce sync.Once
	unregister := func() {
		unregOnce.Do(func() { s.unregisterClient(r, client) })
	}
	defer func() {
		ticker.Stop()
		unregister()
		client.Close()
	}()
	for {
		select {
		case resp, ok := <-client.Send:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := client.Conn.WriteJSON(resp)
			if err != nil {
				return
			}
		case <-ticker.C:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *Server) broadcast(r *room.Room) {
	r.Wsm.Broadcast()
}

func (s *Server) readPump(r2 *room.Room, client *room.WsClient) {
	defer func() {
		s.unregisterClient(r2, client)
		client.Close()
	}()
	for {
		_, msgBytes, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		var wsMsg types.WSMessage
		if err := json.Unmarshal(msgBytes, &wsMsg); err != nil {
			continue
		}

		text := strings.TrimSpace(wsMsg.Text)
		const maxChatRunes = 200
		if runes := []rune(text); len(runes) > maxChatRunes {
			text = string(runes[:maxChatRunes])
		}

		switch wsMsg.Type {
		case "chat":
			if text == "" {
				continue
			}
			var name string
			r2.Sg.Lock()
			for _, p := range r2.Sg.GetGame().Players {
				if p.ID == client.PlayerID {
					name = p.Name
					break
				}
			}
			r2.Sg.Unlock()
			if name == "" {
				if dbUser, _ := store.GetUserByUUID(s.db, client.PlayerID); dbUser != nil {
					name = dbUser.Username
				} else {
					name = "Observer"
				}
			}
			r2.Sg.AddChatMessage(client.PlayerID, name, text)
			s.broadcast(r2)

		case "reaction":
			if text == "" {
				continue
			}
			r2.Sg.Lock()
			var player *game.Player
			for _, p := range r2.Sg.GetGame().Players {
				if p.ID == client.PlayerID {
					player = p
					break
				}
			}
			if player != nil {
				player.Reaction = text
				go func(pID, rx string) {
					time.Sleep(3 * time.Second)
					r2.Sg.Lock()
					for _, p := range r2.Sg.GetGame().Players {
						if p.ID == pID && p.Reaction == rx {
							p.Reaction = ""
							break
						}
					}
					r2.Sg.Unlock()
					s.broadcast(r2)
				}(client.PlayerID, text)
			}
			r2.Sg.Unlock()
			s.broadcast(r2)
		}
	}
}
