package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"lil-poker/internal/card"
	"lil-poker/internal/deck"
	"lil-poker/internal/game"
	"lil-poker/internal/hand"
	"lil-poker/internal/room"
)

func TestHealthEndpoint(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	server := NewServer(db, LoadConfig())
	handler := server.Handler()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to parse health response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", resp["status"])
	}
}

func TestRequireAuthReturns401(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	server := NewServer(db, LoadConfig())

	called := false
	inner := func(w http.ResponseWriter, r *http.Request, u *User) {
		called = true
	}
	handler := server.RequireAuth(inner)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if !called {
		t.Error("in test mode RequireAuth should pass through with nil user")
	}
	_ = w
}

func TestToResponseShowdownCards(t *testing.T) {
	sg := room.NewSafeGame(10, 20)
	sg.Lock()
	g := sg.GetGame()
	if err := g.AddPlayer("p1", "gungnir", 980, 0); err != nil {
		sg.Unlock()
		t.Fatalf("AddPlayer gungnir: %v", err)
	}
	if err := g.AddPlayer("p2", "raime", 20, 1); err != nil {
		sg.Unlock()
		t.Fatalf("AddPlayer raime: %v", err)
	}
	g.Players[0].Hole = deck.New().Deal(2)
	g.Players[1].Hole = deck.New().Deal(2)
	g.Players[1].Folded = true
	g.LastWinners = append(g.LastWinners, game.Winner{Player: g.Players[0], Amount: 20})
	sg.UpdateHandEvaluations()
	sg.Unlock()

	respAnon := sg.GetStatus("", 0)
	if len(respAnon.Players[0].Hole) > 0 {
		t.Error("expected gungnir's hole cards to be hidden from anonymous on fold-out, but got some")
	}

	respGungnir := sg.GetStatus("p1", 0)
	if len(respGungnir.Players[0].Hole) != 2 {
		t.Errorf("expected gungnir to see her own hole cards on fold-out, got %d", len(respGungnir.Players[0].Hole))
	}

	sg2 := room.NewSafeGame(10, 20)
	sg2.Lock()
	g2 := sg2.GetGame()
	_ = g2.AddPlayer("p1", "gungnir", 980, 0)
	_ = g2.AddPlayer("p2", "raime", 20, 1)
	g2.Players[0].Hole = deck.New().Deal(2)
	g2.Players[1].Hole = deck.New().Deal(2)
	g2.Players[0].Folded = false
	g2.Players[1].Folded = false
	winHand := hand.Result{
		Cards: g2.Players[0].Hole,
	}
	g2.LastWinners = append(g2.LastWinners, game.Winner{Player: g2.Players[0], Amount: 20, Hand: winHand})
	sg2.UpdateHandEvaluations()
	sg2.Unlock()

	respAnon2 := sg2.GetStatus("", 0)
	if len(respAnon2.Players[0].Hole) != 2 {
		t.Error("expected gungnir's hole cards to be visible to everyone on showdown, but got none")
	}
	if len(respAnon2.Players[1].Hole) != 2 {
		t.Error("expected raime's hole cards to be visible to everyone on showdown, but got none")
	}
}

func TestPreFlopCall2Players(t *testing.T) {
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

	_ = registerAndSeatPlayer(t, handler, "gungnir", roomID)
	_ = registerAndSeatPlayer(t, handler, "raime", roomID)

	reqStart := httptest.NewRequest("POST", "/api/game/start?room="+roomID, nil)
	wStart := httptest.NewRecorder()
	handler.ServeHTTP(wStart, reqStart)

	reqStatus := httptest.NewRequest("GET", "/api/game/status?room="+roomID, nil)
	wStatus := httptest.NewRecorder()
	handler.ServeHTTP(wStatus, reqStatus)
	var status GameStateResponse
	if err := json.NewDecoder(wStatus.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode game state response: %v", err)
	}

	activeID := status.ActivePlayerID
	if activeID == "" {
		t.Fatalf("expected ActivePlayerID to be set, but it was empty")
	}

	actReq := ActRequest{
		PlayerID: activeID,
		Action:   "call",
	}
	bodyAct, _ := json.Marshal(actReq)
	reqAct := httptest.NewRequest("POST", "/api/game/act?room="+roomID, bytes.NewBuffer(bodyAct))
	wAct := httptest.NewRecorder()
	handler.ServeHTTP(wAct, reqAct)

	if wAct.Code != http.StatusOK {
		t.Fatalf("expected call to succeed, got status %d, body: %s", wAct.Code, wAct.Body.String())
	}

	var statusAfterCall GameStateResponse
	json.NewDecoder(wAct.Body).Decode(&statusAfterCall)

	callerFound := false
	for _, p := range statusAfterCall.Players {
		if p.ID == activeID {
			callerFound = true
			if p.Bet != 20 {
				t.Errorf("expected caller bet to be 20, got %d", p.Bet)
			}
			if !p.Acted {
				t.Errorf("expected caller Acted to be true, got %t", p.Acted)
			}
		}
	}
	if !callerFound {
		t.Fatalf("caller player not found in status response")
	}

	if statusAfterCall.ActivePlayerID == activeID {
		t.Errorf("expected turn to transition to BB, but active player is still caller %s", statusAfterCall.ActivePlayerID)
	}

	otherPlayerID := statusAfterCall.ActivePlayerID
	actReq2 := ActRequest{
		PlayerID: otherPlayerID,
		Action:   "check",
	}
	bodyAct2, _ := json.Marshal(actReq2)
	reqAct2 := httptest.NewRequest("POST", "/api/game/act?room="+roomID, bytes.NewBuffer(bodyAct2))
	wAct2 := httptest.NewRecorder()
	handler.ServeHTTP(wAct2, reqAct2)

	if wAct2.Code != http.StatusOK {
		t.Fatalf("expected check to succeed, got status %d, body: %s", wAct2.Code, wAct2.Body.String())
	}

	var statusAfterCheck GameStateResponse
	json.NewDecoder(wAct2.Body).Decode(&statusAfterCheck)

	if statusAfterCheck.Phase != "Flop" {
		t.Errorf("expected phase to advance to Flop, got %s", statusAfterCheck.Phase)
	}
}

func TestVerifyShowdownHighlight(t *testing.T) {
	sg := room.NewSafeGame(10, 20)
	sg.Lock()
	g := sg.GetGame()
	_ = g.AddPlayer("p1", "gungnir", 980, 0)
	_ = g.AddPlayer("p2", "raime", 20, 1)

	g.Players[0].Hole = []card.Card{
		card.New(card.Ace, card.Spades),
		card.New(card.Nine, card.Clubs),
	}
	g.Players[1].Hole = []card.Card{
		card.New(card.Two, card.Diamonds),
		card.New(card.Three, card.Diamonds),
	}
	g.Board = []card.Card{
		card.New(card.Ace, card.Diamonds),
		card.New(card.Nine, card.Diamonds),
		card.New(card.King, card.Spades),
		card.New(card.Queen, card.Clubs),
		card.New(card.Two, card.Spades),
	}

	gungnirHand := hand.Evaluate(append(g.Players[0].Hole, g.Board...))

	g.LastWinners = append(g.LastWinners, game.Winner{
		Player: g.Players[0],
		Amount: 40,
		Hand:   gungnirHand,
	})

	sg.UpdateHandEvaluations()
	sg.Unlock()

	status := sg.GetStatus("p1", 0)

	if len(status.LastWinners) != 1 {
		t.Fatalf("expected 1 winner, got %d", len(status.LastWinners))
	}

	winner := status.LastWinners[0]
	if winner.HandRank != "Two Pair, Aces and Nines" {
		t.Errorf("expected HandRank 'Two Pair, Aces and Nines', got %q", winner.HandRank)
	}

	expectedCards := map[string]bool{
		"A♠": true,
		"A♦": true,
		"9♣": true,
		"9♦": true,
	}

	if len(winner.HandCards) != 4 {
		t.Errorf("expected 4 combination cards, got %d: %v", len(winner.HandCards), winner.HandCards)
	}

	for _, c := range winner.HandCards {
		cardStr := c.String()
		if !expectedCards[cardStr] {
			t.Errorf("unexpected card %s in combination cards", cardStr)
		}
	}
}

func TestGuestResumePassword(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	server := NewServer(db, LoadConfig())
	handler := server.Handler()

	guestReq := map[string]string{"username": "testguest"}
	body, _ := json.Marshal(guestReq)
	req := httptest.NewRequest("POST", "/api/auth/guest", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode guest response: %v", err)
	}

	uuid, _ := resp["uuid"].(string)
	password, _ := resp["password"].(string)

	if uuid == "" || password == "" {
		t.Fatalf("expected non-empty uuid and password, got uuid=%q, password=%q", uuid, password)
	}

	resumeReq := map[string]string{"uuid": uuid, "password": password}
	bodyResume, _ := json.Marshal(resumeReq)
	reqResume := httptest.NewRequest("POST", "/api/auth/guest/resume", bytes.NewBuffer(bodyResume))
	wResume := httptest.NewRecorder()
	handler.ServeHTTP(wResume, reqResume)

	if wResume.Code != http.StatusOK {
		t.Errorf("expected 200 on valid guest resume, got %d. Body: %s", wResume.Code, wResume.Body.String())
	}

	resumeReqWrongPass := map[string]string{"uuid": uuid, "password": "wrongpassword"}
	bodyResumeWP, _ := json.Marshal(resumeReqWrongPass)
	reqResumeWP := httptest.NewRequest("POST", "/api/auth/guest/resume", bytes.NewBuffer(bodyResumeWP))
	wResumeWP := httptest.NewRecorder()
	handler.ServeHTTP(wResumeWP, reqResumeWP)

	if wResumeWP.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on guest resume with wrong password, got %d", wResumeWP.Code)
	}

	resumeReqWrongUUID := map[string]string{"uuid": "00000000-0000-0000-0000-000000000000", "password": password}
	bodyResumeWU, _ := json.Marshal(resumeReqWrongUUID)
	reqResumeWU := httptest.NewRequest("POST", "/api/auth/guest/resume", bytes.NewBuffer(bodyResumeWU))
	wResumeWU := httptest.NewRecorder()
	handler.ServeHTTP(wResumeWU, reqResumeWU)

	if wResumeWU.Code != http.StatusNotFound {
		t.Errorf("expected 404 on guest resume with wrong uuid, got %d", wResumeWU.Code)
	}
}

func TestCSRFOriginChecking(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	cfg := LoadConfig()
	cfg.AllowedOrigins = []string{"http://allowed.com"}
	server := NewServer(db, cfg)
	handler := server.Handler()

	body, _ := json.Marshal(AuthRequest{Username: "csrfuser1", Password: "testpassword"})
	reqOk := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	reqOk.Header.Set("Origin", "http://allowed.com")
	wOk := httptest.NewRecorder()
	handler.ServeHTTP(wOk, reqOk)

	if wOk.Code != http.StatusCreated {
		t.Errorf("expected 201 Created from allowed origin, got %d. Body: %s", wOk.Code, wOk.Body.String())
	}

	body2, _ := json.Marshal(AuthRequest{Username: "csrfuser2", Password: "testpassword"})
	reqBad := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body2))
	reqBad.Header.Set("Origin", "http://malicious.com")
	wBad := httptest.NewRecorder()
	handler.ServeHTTP(wBad, reqBad)

	if wBad.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden from malicious origin, got %d", wBad.Code)
	}
}

func TestCustomizableBlindEscalation(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	server := NewServer(db, LoadConfig())
	handler := server.Handler()

	body1 := []byte(`{"name":"15 Mins Room","small_blind":10,"big_blind":20,"blind_escalation_mins":15}`)
	req1 := httptest.NewRequest("POST", "/api/rooms", bytes.NewBuffer(body1))
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("failed to create room: status %d", w1.Code)
	}
	var ri1 RoomInfo
	json.NewDecoder(w1.Body).Decode(&ri1)

	if ri1.BlindEscalationMins != 15 {
		t.Errorf("expected blind_escalation_mins to be 15, got %d", ri1.BlindEscalationMins)
	}

	body2 := []byte(`{"name":"No Escalation Room","small_blind":10,"big_blind":20,"blind_escalation_mins":0}`)
	req2 := httptest.NewRequest("POST", "/api/rooms", bytes.NewBuffer(body2))
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("failed to create room: status %d", w2.Code)
	}
	var ri2 RoomInfo
	json.NewDecoder(w2.Body).Decode(&ri2)

	if ri2.BlindEscalationMins != 0 {
		t.Errorf("expected blind_escalation_mins to be 0, got %d", ri2.BlindEscalationMins)
	}
}