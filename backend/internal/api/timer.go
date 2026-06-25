package api

import (
	"log/slog"
	"time"

	"lil-poker/internal/game"
	"lil-poker/internal/room"
)

func (s *Server) resetTimer(r *room.Room) {
	r.TimerMu.Lock()
	defer r.TimerMu.Unlock()

	if r.TimerCancel != nil {
		close(r.TimerCancel)
		r.TimerCancel = nil
	}

	timeout := s.turnTimeout
	if r.TurnTimeoutSecs > 0 {
		timeout = time.Duration(r.TurnTimeoutSecs) * time.Second
	}

	r.Sg.Lock()
	g := r.Sg.GetGame()
	isActive := g.Phase != game.PhaseWaiting && g.ActiveIdx >= 0 && g.ActiveIdx < len(g.Players)

	var activeID string
	if isActive {
		activeID = g.Players[g.ActiveIdx].ID
		r.Sg.SetActionDeadline(time.Now().Add(timeout).UnixMilli())
	} else {
		r.Sg.SetActionDeadline(0)
	}
	r.Sg.Unlock()

	if !isActive {
		return
	}

	r.TimerCancel = make(chan struct{})
	cancelChan := r.TimerCancel

	go func(pID string, ch chan struct{}, t time.Duration) {
		select {
		case <-ch:
			return
		case <-time.After(t):
			s.HandleTurnTimeout(r, pID)
		}
	}(activeID, cancelChan, timeout)
}

func (s *Server) HandleTurnTimeout(r *room.Room, playerID string) {
	r.Sg.Lock()
	g := r.Sg.GetGame()

	if g.Phase == game.PhaseWaiting || g.ActiveIdx < 0 || g.ActiveIdx >= len(g.Players) {
		r.Sg.Unlock()
		return
	}
	activePlayer := g.Players[g.ActiveIdx]
	if activePlayer.ID != playerID {
		r.Sg.Unlock()
		return
	}

	activePlayer.TimeoutCount++
	autoSitOut := activePlayer.TimeoutCount >= 2

	action := game.ActionFold
	if activePlayer.Bet == g.CurrentBet {
		action = game.ActionCheck
	}

	slog.Info("Player timed out", "room_id", r.ID, "player", activePlayer.Name, "action", actionStr(action))

	err := g.Act(playerID, action, 0)
	var chipUpdates map[string]int
	if err == nil {
		r.Sg.UpdateHandEvaluations()
		r.Sg.CheckRecordHandHistoryLocked()
		if g.Phase == game.PhaseWaiting {
			chipUpdates = make(map[string]int, len(g.Players))
			for _, p := range g.Players {
				chipUpdates[p.ID] = p.Chips
			}
		}
	}

	if autoSitOut {
		slog.Warn("Player auto sit-out due to timeouts", "room_id", r.ID, "player", activePlayer.Name, "timeouts", activePlayer.TimeoutCount)
		_ = g.SitOut(playerID)
	}
	r.Sg.Unlock()

	if r.StartingChips == 0 {
		s.flushChipsAsync(chipUpdates)
	}
	s.broadcast(r)
	s.resetTimer(r)
}

func actionStr(a game.Action) string {
	if a == game.ActionCheck {
		return "check"
	}
	return "fold"
}
