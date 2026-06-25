package room

import (
	"lil-poker/internal/card"
	"lil-poker/internal/game"
	"lil-poker/internal/types"
)

type playerSnapshot struct {
	ID            string
	Name          string
	Chips         int
	Hole          []card.Card
	Bet           int
	Folded        bool
	AllIn         bool
	Acted         bool
	HandStrength  string
	SittingOut    bool
	Seat          int
	Reaction      string
	HandsPlayed   int
	HandsVPIP     int
	BiggestPotWon int
	IsSmallBlind  bool
	IsBigBlind    bool
}

type gameSnapshot struct {
	Players             []playerSnapshot
	Board               []card.Card
	Phase               string
	Pot                 int
	CurrentBet          int
	ActiveIdx           int
	ActiveName          string
	ActivePlayerID      string
	DealerIdx           int
	SmallBlind          int
	BigBlind            int
	LastWinners         []types.WinnerStatus
	ActionDeadline      int64
	HandCount           int
	ChatMessages        []types.ChatMessage
	NextSmallBlind      int
	NextBigBlind        int
	HandsUntilRaise     int
	BlindsRaiseDeadline int64
	HandHistory         []types.HandHistoryEntry
	CreatorID           string
	SubPots             []types.SubPot
}

func (sg *SafeGame) getSnapshot() gameSnapshot {
	sg.mu.RLock()
	defer sg.mu.RUnlock()

	g := sg.game

	players := make([]playerSnapshot, len(g.Players))
	for i, p := range g.Players {
		holeCopy := make([]card.Card, len(p.Hole))
		copy(holeCopy, p.Hole)
		players[i] = playerSnapshot{
			ID:            p.ID,
			Name:          p.Name,
			Chips:         p.Chips,
			Hole:          holeCopy,
			Bet:           p.Bet,
			Folded:        p.Folded,
			AllIn:         p.AllIn,
			Acted:         p.Acted,
			HandStrength:  p.HandStrength,
			SittingOut:    p.SittingOut,
			Seat:          p.Seat,
			Reaction:      p.Reaction,
			HandsPlayed:   p.HandsPlayed,
			HandsVPIP:     p.HandsVPIP,
			BiggestPotWon: p.BiggestPotWon,
			IsSmallBlind:  i == g.SmallBlindIdx,
			IsBigBlind:    i == g.BigBlindIdx,
		}
	}

	boardCopy := make([]card.Card, len(g.Board))
	copy(boardCopy, g.Board)

	var activeName string
	if len(g.Players) > 0 && g.ActiveIdx >= 0 && g.ActiveIdx < len(g.Players) {
		activeName = g.Players[g.ActiveIdx].Name
	}

	var winners []types.WinnerStatus
	if len(g.LastWinners) > 0 {
		winners = make([]types.WinnerStatus, len(g.LastWinners))
		for i, w := range g.LastWinners {
			ws := types.WinnerStatus{
				PlayerID:   w.Player.ID,
				PlayerName: w.Player.Name,
				Amount:     w.Amount,
			}
			if len(w.Hand.Cards) > 0 {
				ws.HandRank = w.Hand.Description()
				ws.HandCards = w.Hand.GetCombinationCards()
			}
			winners[i] = ws
		}
	}

	var nextSB, nextBB, handsLeft int
	if g.HandCount > 0 {
		nextRaiseHand := ((g.HandCount / 10) + 1) * 10
		handsLeft = nextRaiseHand - g.HandCount
		nextSB, nextBB = g.NextBlindLevel()
	} else {
		handsLeft = 10
		nextSB, nextBB = g.NextBlindLevel()
	}

	chatCopy := make([]types.ChatMessage, len(sg.chatMessages))
	copy(chatCopy, sg.chatMessages)

	var blindsRaiseDeadline int64
	if !g.NextRaiseTime.IsZero() {
		blindsRaiseDeadline = g.NextRaiseTime.UnixMilli()
	}

	historyCopy := make([]types.HandHistoryEntry, len(sg.handHistory))
	copy(historyCopy, sg.handHistory)

	gSubPots := g.CalculateCurrentSubPots()
	subPots := make([]types.SubPot, len(gSubPots))
	for i, sp := range gSubPots {
		subPots[i] = types.SubPot{
			Amount:       sp.Amount,
			Contributors: sp.Contributors,
		}
	}

	return gameSnapshot{
		Players:             players,
		Board:               boardCopy,
		Phase:               g.Phase.String(),
		Pot:                 g.Pot,
		CurrentBet:          g.CurrentBet,
		ActiveIdx:           g.ActiveIdx,
		ActiveName:          activeName,
		ActivePlayerID:      g.ActivePlayerID,
		DealerIdx:           g.DealerIdx,
		SmallBlind:          g.SmallBlind,
		BigBlind:            g.BigBlind,
		LastWinners:         winners,
		ActionDeadline:      sg.actionDeadline.Load(),
		HandCount:           g.HandCount,
		ChatMessages:        chatCopy,
		NextSmallBlind:      nextSB,
		NextBigBlind:        nextBB,
		HandsUntilRaise:     handsLeft,
		BlindsRaiseDeadline: blindsRaiseDeadline,
		HandHistory:         historyCopy,
		CreatorID:           sg.CreatorID,
		SubPots:             subPots,
	}
}

func (snap gameSnapshot) toResponse(requestingPlayerID string, observerCount int) types.GameStateResponse {
	players := make([]types.PlayerStatus, len(snap.Players))
	for i, p := range snap.Players {
		ps := types.PlayerStatus{
			ID:            p.ID,
			Name:          p.Name,
			Chips:         p.Chips,
			Bet:           p.Bet,
			Folded:        p.Folded,
			AllIn:         p.AllIn,
			Acted:         p.Acted,
			SittingOut:    p.SittingOut,
			Seat:          p.Seat,
			Reaction:      p.Reaction,
			HandsPlayed:   p.HandsPlayed,
			HandsVPIP:     p.HandsVPIP,
			BiggestPotWon: p.BiggestPotWon,
			IsSmallBlind:  p.IsSmallBlind,
			IsBigBlind:    p.IsBigBlind,
		}

		showCards := false
		if snap.Phase == "Waiting" {
			hasShowdown := false
			for _, w := range snap.LastWinners {
				if len(w.HandCards) > 0 {
					hasShowdown = true
					break
				}
			}

			if hasShowdown {
				if !p.Folded {
					showCards = true
				}
			} else {
				if requestingPlayerID == p.ID {
					showCards = true
				}
			}
		} else {
			if requestingPlayerID == p.ID {
				showCards = true
			}
		}

		if showCards && len(p.Hole) > 0 {
			ps.Hole = p.Hole
		}

		if requestingPlayerID == p.ID && len(p.Hole) > 0 && snap.Phase != "Waiting" && !p.Folded {
			ps.CurrentHand = p.HandStrength
		}

		players[i] = ps
	}

	return types.GameStateResponse{
		Players:             players,
		Board:               snap.Board,
		Phase:               snap.Phase,
		Pot:                 snap.Pot,
		CurrentBet:          snap.CurrentBet,
		ActiveIdx:           snap.ActiveIdx,
		ActiveName:          snap.ActiveName,
		ActivePlayerID:      snap.ActivePlayerID,
		DealerIdx:           snap.DealerIdx,
		SmallBlind:          snap.SmallBlind,
		BigBlind:            snap.BigBlind,
		LastWinners:         snap.LastWinners,
		ActionDeadline:      snap.ActionDeadline,
		HandCount:           snap.HandCount,
		ChatMessages:        snap.ChatMessages,
		NextSmallBlind:      snap.NextSmallBlind,
		NextBigBlind:        snap.NextBigBlind,
		HandsUntilRaise:     snap.HandsUntilRaise,
		ObserverCount:       observerCount,
		BlindsRaiseDeadline: snap.BlindsRaiseDeadline,
		HandHistory:         snap.HandHistory,
		CreatorID:           snap.CreatorID,
		SubPots:             snap.SubPots,
	}
}

func (sg *SafeGame) GetStatus(requestingPlayerID string, observerCount int) types.GameStateResponse {
	snap := sg.getSnapshot()
	return snap.toResponse(requestingPlayerID, observerCount)
}

func (sg *SafeGame) GetGame() *game.Game {
	return sg.game
}


func (sg *SafeGame) Lock() {
	sg.mu.Lock()
}

func (sg *SafeGame) Unlock() {
	sg.mu.Unlock()
}

func (sg *SafeGame) RLock() {
	sg.mu.RLock()
}

func (sg *SafeGame) RUnlock() {
	sg.mu.RUnlock()
}
