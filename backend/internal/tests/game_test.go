package tests

import (
	"testing"

	"lil-poker/internal/game"
)

func TestNextBlindLevel_Doubles(t *testing.T) {
	g := game.NewGame(10, 20)
	g.Players = append(g.Players, &game.Player{ID: "p1", Name: "gungnir", Chips: 10000})

	sb, bb := g.NextBlindLevel()
	if sb != 20 || bb != 40 {
		t.Errorf("expected next level 20/40, got %d/%d", sb, bb)
	}
}

func TestNextBlindLevel_Capped(t *testing.T) {
	g := game.NewGame(50, 100)
	g.Players = append(g.Players, &game.Player{ID: "p1", Name: "gungnir", Chips: 120})

	sb, bb := g.NextBlindLevel()
	if sb != g.SmallBlind || bb != g.BigBlind {
		t.Errorf("expected capped blinds %d/%d, got %d/%d", g.SmallBlind, g.BigBlind, sb, bb)
	}
}

func TestNextBlindLevel_NoPlayers(t *testing.T) {
	g := game.NewGame(10, 20)
	sb, bb := g.NextBlindLevel()
	if sb != 20 || bb != 40 {
		t.Errorf("expected 20/40 with no players, got %d/%d", sb, bb)
	}
}

func TestNextBlindLevel_AllSittingOut(t *testing.T) {
	g := game.NewGame(10, 20)
	g.Players = append(g.Players, &game.Player{ID: "p1", Name: "gungnir", Chips: 500, SittingOut: true})
	sb, bb := g.NextBlindLevel()
	if sb != 20 || bb != 40 {
		t.Errorf("expected 20/40 when all sitting out, got %d/%d", sb, bb)
	}
}

func TestAddPlayer_DuplicateID(t *testing.T) {
	g := game.NewGame(10, 20)
	if err := g.AddPlayer("u1", "gungnir", 1000, 0); err != nil {
		t.Fatalf("first add should succeed: %v", err)
	}
	err := g.AddPlayer("u1", "raime", 1000, 1)
	if err == nil {
		t.Error("expected error for duplicate player ID, got nil")
	}
}

func TestAddPlayer_ZeroChips(t *testing.T) {
	g := game.NewGame(10, 20)
	err := g.AddPlayer("u1", "Broke", 0, 0)
	if err == nil {
		t.Error("expected error for 0 chips, got nil")
	}
}

func TestAddPlayer_TableFull(t *testing.T) {
	g := game.NewGame(10, 20)
	for i := 0; i < 8; i++ {
		id := string(rune('a' + i))
		name := string(rune('A' + i))
		if err := g.AddPlayer(id, name, 1000, i); err != nil {
			t.Fatalf("adding player %d failed: %v", i, err)
		}
	}
	err := g.AddPlayer("z", "Extra", 1000, -1)
	if err == nil {
		t.Error("expected error when table is full, got nil")
	}
}

func TestRemovePlayer(t *testing.T) {
	g := game.NewGame(10, 20)
	_ = g.AddPlayer("p1", "gungnir", 1000, 0)
	_ = g.AddPlayer("p2", "raime", 1000, 1)
	_ = g.AddPlayer("p3", "player3", 1000, 2)

	if len(g.Players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(g.Players))
	}

	err := g.RemovePlayer("p3")
	if err != nil {
		t.Fatalf("failed to remove p3: %v", err)
	}

	if len(g.Players) != 2 {
		t.Errorf("expected 2 players remaining, got %d", len(g.Players))
	}

	for _, p := range g.Players {
		if p.ID == "p3" {
			t.Error("player p3 should have been removed")
		}
	}
}
