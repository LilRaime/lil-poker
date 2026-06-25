package room

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"lil-poker/internal/card"
	"lil-poker/internal/game"
	"lil-poker/internal/hand"
	"lil-poker/internal/types"
)

type SafeGame struct {
	mu                    sync.RWMutex
	game                  *game.Game
	actionDeadline        atomic.Int64
	chatMessages          []types.ChatMessage
	handHistory           []types.HandHistoryEntry
	lastRecordedHandCount int
	CreatorID             string
}

func NewSafeGame(sb, bb int) *SafeGame {
	return &SafeGame{
		game: game.NewGame(sb, bb),
	}
}

func (sg *SafeGame) CheckRecordHandHistoryLocked() {
	g := sg.game
	if g.Phase == game.PhaseWaiting && len(g.LastWinners) > 0 && g.HandCount > sg.lastRecordedHandCount {
		sg.lastRecordedHandCount = g.HandCount

		winners := make([]types.WinnerStatus, len(g.LastWinners))
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

		entry := types.HandHistoryEntry{
			HandNum: g.HandCount,
			Board:   make([]card.Card, len(g.Board)),
			Winners: winners,
		}
		copy(entry.Board, g.Board)

		const maxHistory = 20
		if len(sg.handHistory) < maxHistory {
			sg.handHistory = append(sg.handHistory, entry)
		} else {
			copy(sg.handHistory, sg.handHistory[1:])
			sg.handHistory[maxHistory-1] = entry
		}
	}
}

func (sg *SafeGame) CheckRecordHandHistory() {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	sg.CheckRecordHandHistoryLocked()
}

func (sg *SafeGame) addMessage(name, playerID, text string, system bool) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	const maxMessages = 50
	msg := types.ChatMessage{
		PlayerName: name,
		PlayerID:   playerID,
		Text:       text,
		Time:       time.Now().UnixMilli(),
		System:     system,
	}
	if len(sg.chatMessages) < maxMessages {
		sg.chatMessages = append(sg.chatMessages, msg)
	} else {
		copy(sg.chatMessages, sg.chatMessages[1:])
		sg.chatMessages[maxMessages-1] = msg
	}
}

func (sg *SafeGame) AddChatMessage(senderID, senderName, text string) {
	sg.addMessage(senderName, senderID, text, false)
}

func (sg *SafeGame) AddSystemMessage(text string) {
	sg.addMessage("System", "", text, true)
}

func (sg *SafeGame) GetActionDeadline() int64 {
	return sg.actionDeadline.Load()
}

func (sg *SafeGame) SetActionDeadline(d int64) {
	sg.actionDeadline.Store(d)
}

func higherRank(r1, r2 card.Rank) card.Rank {
	if r1 > r2 {
		return r1
	}
	return r2
}

func (sg *SafeGame) UpdateHandEvaluations() {
	g := sg.game
	if g.Phase == game.PhaseWaiting {
		return
	}
	for _, p := range g.Players {
		if p.Folded || len(p.Hole) == 0 {
			p.HandStrength = ""
			continue
		}
		combined := append([]card.Card{}, p.Hole...)
		combined = append(combined, g.Board...)
		if len(combined) >= 2 {
			if len(combined) == 2 {
				if p.Hole[0].Rank == p.Hole[1].Rank {
					p.HandStrength = fmt.Sprintf("Pocket Pair of %ss", rankPluralString(p.Hole[0].Rank))
				} else {
					p.HandStrength = fmt.Sprintf("High Card %s", rankLongString(higherRank(p.Hole[0].Rank, p.Hole[1].Rank)))
				}
			} else {
				res := hand.Evaluate(combined)
				p.HandStrength = getHandDescription(res)
			}
		} else {
			p.HandStrength = ""
		}
	}
}
