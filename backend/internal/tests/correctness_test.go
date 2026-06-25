package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"lil-poker/internal/game"
	"lil-poker/internal/room"
)

func TestHeadsUpBlinds(t *testing.T) {
	g := game.NewGame(10, 20)
	_ = g.AddPlayer("p1", "Player 1", 1000, 0)
	_ = g.AddPlayer("p2", "Player 2", 1000, 1)

	err := g.Start()
	if err != nil {
		t.Fatalf("failed to start game: %v", err)
	}

	dealerIdx := g.DealerIdx
	sbIdx := g.SmallBlindIdx
	bbIdx := g.BigBlindIdx

	if sbIdx != dealerIdx {
		t.Errorf("expected Small Blind to be the dealer (%d), got %d", dealerIdx, sbIdx)
	}

	expectedBBIdx := (dealerIdx + 1) % 2
	if bbIdx != expectedBBIdx {
		t.Errorf("expected Big Blind to be non-dealer (%d), got %d", expectedBBIdx, bbIdx)
	}
}

func TestMinRaiseEnforcement(t *testing.T) {
	g := game.NewGame(10, 20)
	_ = g.AddPlayer("p1", "Player 1", 1000, 0)
	_ = g.AddPlayer("p2", "Player 2", 1000, 1)

	g.HandCount = 0
	err := g.Start()
	if err != nil {
		t.Fatalf("failed to start game: %v", err)
	}

	activeIdx := g.ActiveIdx
	activePlayer := g.Players[activeIdx]

	err = g.Act(activePlayer.ID, game.ActionRaise, 30)
	if err == nil {
		t.Error("expected error for sub-min raise, got nil")
	}

	err = g.Act(activePlayer.ID, game.ActionRaise, 40)
	if err != nil {
		t.Fatalf("valid raise to 40 failed: %v", err)
	}

	activeIdx = g.ActiveIdx
	activePlayer = g.Players[activeIdx]

	err = g.Act(activePlayer.ID, game.ActionRaise, 50)
	if err == nil {
		t.Error("expected error for sub-min raise to 50, got nil")
	}

	activePlayer.Chips = 15
	err = g.Act(activePlayer.ID, game.ActionRaise, 35)
	if err != nil {
		t.Fatalf("expected all-in raise to 35 to succeed, got: %v", err)
	}
	if !activePlayer.AllIn {
		t.Error("expected player to be marked all-in")
	}
}

func TestSidePotsCalculation(t *testing.T) {
	g := game.NewGame(10, 20)
	_ = g.AddPlayer("p1", "Player 1", 100, 0)
	_ = g.AddPlayer("p2", "Player 2", 300, 1)
	_ = g.AddPlayer("p3", "Player 3", 500, 2)
	_ = g.AddPlayer("p4", "Player 4", 200, 3)

	g.Players[0].TotalContributed = 100
	g.Players[0].Folded = false
	g.Players[1].TotalContributed = 300
	g.Players[1].Folded = false
	g.Players[2].TotalContributed = 500
	g.Players[2].Folded = false
	g.Players[3].TotalContributed = 200
	g.Players[3].Folded = true

	g.Phase = game.PhaseFlop

	subPots := g.CalculateCurrentSubPots()
	if len(subPots) != 3 {
		t.Fatalf("expected 3 sub pots, got %d", len(subPots))
	}

	if subPots[0].Amount != 400 {
		t.Errorf("expected main pot amount 400, got %d", subPots[0].Amount)
	}
	if len(subPots[0].Contributors) != 3 {
		t.Errorf("expected 3 contributors for main pot, got %d", len(subPots[0].Contributors))
	}

	if subPots[1].Amount != 500 {
		t.Errorf("expected side pot 1 amount 500, got %d", subPots[1].Amount)
	}
	if len(subPots[1].Contributors) != 2 {
		t.Errorf("expected 2 contributors for side pot 1, got %d", len(subPots[1].Contributors))
	}

	if subPots[2].Amount != 200 {
		t.Errorf("expected side pot 2 amount 200, got %d", subPots[2].Amount)
	}
	if len(subPots[2].Contributors) != 1 || subPots[2].Contributors[0] != "Player 3" {
		t.Errorf("expected Player 3 as contributor for side pot 2, got %v", subPots[2].Contributors)
	}
}

func TestConnectionCollision(t *testing.T) {
	sg := room.NewSafeGame(10, 20)
	wsm := room.NewWSManager(sg)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade failed: %v", err)
			return
		}
		client := room.NewWsClient(conn, "player1", "User1")
		wsm.Register(client)

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer s.Close()

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	if !wsm.IsPlayerConnected("player1") {
		t.Error("expected player1 to be connected")
	}

	wsm.DisconnectPlayer("player1")

	if wsm.IsPlayerConnected("player1") {
		t.Error("expected player1 to be disconnected after collision")
	}
}
