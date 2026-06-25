package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPokerAPIWorkflow(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	server := NewServer(db, LoadConfig())
	handler := server.Handler()

	roomID := createTestRoom(t, handler)

	createReq := CreateGameRequest{SmallBlind: 5, BigBlind: 10}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/game/create?room="+roomID, bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	_, _, _ = registerAndSeatPlayer(t, handler, "gungnir", roomID),
		registerAndSeatPlayer(t, handler, "raime", roomID),
		registerAndSeatPlayer(t, handler, "player3", roomID)

	reqStatus := httptest.NewRequest("GET", "/api/game/status?room="+roomID, nil)
	wStatus := httptest.NewRecorder()
	handler.ServeHTTP(wStatus, reqStatus)

	var statusRes GameStateResponse
	if err := json.NewDecoder(wStatus.Body).Decode(&statusRes); err != nil {
		t.Fatalf("failed to decode status response: %v", err)
	}
	if len(statusRes.Players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(statusRes.Players))
	}
	if statusRes.Phase != "Waiting" {
		t.Fatalf("expected phase to be Waiting, got %s", statusRes.Phase)
	}

	reqStart := httptest.NewRequest("POST", "/api/game/start?room="+roomID, nil)
	wStart := httptest.NewRecorder()
	handler.ServeHTTP(wStart, reqStart)

	if wStart.Code != http.StatusOK {
		t.Fatalf("failed to start game: status %d, body: %s", wStart.Code, wStart.Body.String())
	}

	reqStatusAnon := httptest.NewRequest("GET", "/api/game/status?room="+roomID, nil)
	wStatusAnon := httptest.NewRecorder()
	handler.ServeHTTP(wStatusAnon, reqStatusAnon)
	var statusAnon GameStateResponse
	if err := json.NewDecoder(wStatusAnon.Body).Decode(&statusAnon); err != nil {
		t.Fatalf("failed to decode anonymous status response: %v", err)
	}

	for _, p := range statusAnon.Players {
		if len(p.Hole) > 0 {
			t.Errorf("expected hole cards to be hidden for anonymous viewer, but player %s has cards %v", p.Name, p.Hole)
		}
	}

	activePlayerID := ""
	for _, p := range statusAnon.Players {
		if p.Name == statusAnon.ActiveName {
			activePlayerID = p.ID
			break
		}
	}

	if activePlayerID == "" {
		t.Fatalf("could not determine active player ID")
	}

	actReq := ActRequest{
		PlayerID: activePlayerID,
		Action:   "call",
	}
	bodyAct, _ := json.Marshal(actReq)
	reqAct := httptest.NewRequest("POST", "/api/game/act?room="+roomID, bytes.NewBuffer(bodyAct))
	wAct := httptest.NewRecorder()
	handler.ServeHTTP(wAct, reqAct)

	if wAct.Code != http.StatusOK {
		t.Fatalf("expected status 200 on action, got %d. Body: %s", wAct.Code, wAct.Body.String())
	}
}

func TestAllInFlow(t *testing.T) {
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

	reqStatus := httptest.NewRequest("GET", "/api/game/status?room="+roomID, nil)
	wStatus := httptest.NewRecorder()
	handler.ServeHTTP(wStatus, reqStatus)
	var status GameStateResponse
	if err := json.NewDecoder(wStatus.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode status response: %v", err)
	}

	activeID := p1ID
	inactiveID := p2ID
	if status.ActiveName == "raime" {
		activeID = p2ID
		inactiveID = p1ID
	}

	act1 := ActRequest{PlayerID: activeID, Action: "all_in"}
	bodyAct1, _ := json.Marshal(act1)
	reqAct1 := httptest.NewRequest("POST", "/api/game/act?room="+roomID, bytes.NewBuffer(bodyAct1))
	wAct1 := httptest.NewRecorder()
	handler.ServeHTTP(wAct1, reqAct1)
	if wAct1.Code != http.StatusOK {
		t.Fatalf("expected all_in to succeed, got status %d, body: %s", wAct1.Code, wAct1.Body.String())
	}

	act2 := ActRequest{PlayerID: inactiveID, Action: "call"}
	bodyAct2, _ := json.Marshal(act2)
	reqAct2 := httptest.NewRequest("POST", "/api/game/act?room="+roomID, bytes.NewBuffer(bodyAct2))
	wAct2 := httptest.NewRecorder()
	handler.ServeHTTP(wAct2, reqAct2)
	if wAct2.Code != http.StatusOK {
		t.Fatalf("expected call to succeed, got status %d, body: %s", wAct2.Code, wAct2.Body.String())
	}

	var finalStatus GameStateResponse
	if err := json.NewDecoder(wAct2.Body).Decode(&finalStatus); err != nil {
		t.Fatalf("failed to decode final status: %v", err)
	}

	if finalStatus.Phase != "Waiting" {
		t.Errorf("expected game to finish and return to Waiting phase, got %s", finalStatus.Phase)
	}
	if finalStatus.Pot != 0 {
		t.Errorf("expected pot to be awarded and become 0, got %d", finalStatus.Pot)
	}
	if len(finalStatus.LastWinners) == 0 {
		t.Errorf("expected at least one winner, got 0")
	}
}
