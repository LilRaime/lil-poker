package tests

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestTurnTimeout(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	server := NewServer(db, LoadConfig())
	handler := server.Handler()

	roomID := createTestRoom(t, handler)

	createReq := CreateGameRequest{SmallBlind: 10, BigBlind: 20}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/game/create?room="+roomID, bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	p1ID := registerAndSeatPlayer(t, handler, "gungnir", roomID)
	p2ID := registerAndSeatPlayer(t, handler, "raime", roomID)

	reqStart := httptest.NewRequest("POST", "/api/game/start?room="+roomID, nil)
	wStart := httptest.NewRecorder()
	handler.ServeHTTP(wStart, reqStart)

	room, _ := server.RoomManager().GetRoom(roomID)
	room.Sg.Lock()
	g := room.Sg.GetGame()
	activePlayerID := g.Players[g.ActiveIdx].ID
	room.Sg.Unlock()

	if activePlayerID != p1ID && activePlayerID != p2ID {
		t.Fatalf("invalid active player ID")
	}

	server.HandleTurnTimeout(room, activePlayerID)

	room.Sg.Lock()
	p := room.Sg.GetGame().Players[0]
	if p.ID == activePlayerID {
		if !p.Folded {
			t.Errorf("expected player to be folded on timeout, but folded is false")
		}
	} else {
		p2 := room.Sg.GetGame().Players[1]
		if p2.ID == activePlayerID {
			if !p2.Folded {
				t.Errorf("expected player to be folded on timeout, but folded is false")
			}
		}
	}
	room.Sg.Unlock()
}
