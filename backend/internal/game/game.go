package game

import (
	"errors"
	"fmt"
	"log/slog"
	"lil-poker/internal/card"
	"lil-poker/internal/deck"
	"lil-poker/internal/hand"
	"math/rand"
	"strings"
	"time"
)

type Phase int

const (
	PhaseWaiting Phase = iota
	PhasePreFlop
	PhaseFlop
	PhaseTurn
	PhaseRiver
	PhaseShowdown
)

func (p Phase) String() string {
	return [...]string{"Waiting", "Pre-Flop", "Flop", "Turn", "River", "Showdown"}[p]
}

type Action int

const (
	ActionFold Action = iota
	ActionCheck
	ActionCall
	ActionRaise
	ActionAllIn
)

type Player struct {
	ID               string
	Name             string
	Chips            int
	Hole             []card.Card
	Bet              int
	TotalContributed int
	Folded           bool
	AllIn            bool
	Acted            bool
	HandStrength     string
	TimeoutCount     int
	SittingOut       bool
	Seat             int
	Reaction         string
	HandsPlayed      int
	HandsVPIP        int
	BiggestPotWon    int
	IsVPIPThisHand   bool
	RebuysRemaining  int
	ExposedCards     bool
}

func (p *Player) String() string {
	return fmt.Sprintf("%s (chips: %d, bet: %d, seat: %d)", p.Name, p.Chips, p.Bet, p.Seat)
}

type Game struct {
	Players             []*Player
	Board               []card.Card
	Deck                *deck.Deck
	Phase               Phase
	Pot                 int
	CurrentBet          int
	ActiveIdx           int
	ActivePlayerID      string
	DealerIdx           int
	SmallBlind          int
	BigBlind            int
	LastWinners         []Winner
	HandCount           int
	NextRaiseTime       time.Time
	BlindEscalationMins int
	SmallBlindIdx       int
	BigBlindIdx         int
	LastRaiseSize       int
	MaxRebuys           int
}

func NewGame(smallBlind, bigBlind int) *Game {
	return &Game{
		SmallBlind:    smallBlind,
		BigBlind:      bigBlind,
		Phase:         PhaseWaiting,
		SmallBlindIdx: -1,
		BigBlindIdx:   -1,
		LastRaiseSize: bigBlind,
	}
}

func (g *Game) Reset(smallBlind, bigBlind int) {
	g.SmallBlind = smallBlind
	g.BigBlind = bigBlind
	g.Phase = PhaseWaiting
	g.Board = make([]card.Card, 0)
	g.Pot = 0
	g.CurrentBet = 0
	g.ActiveIdx = -1
	g.ActivePlayerID = ""
	g.LastWinners = nil
	g.HandCount = 0
	g.NextRaiseTime = time.Time{}
	g.SmallBlindIdx = -1
	g.BigBlindIdx = -1
	g.LastRaiseSize = bigBlind
	for _, p := range g.Players {
		p.Hole = make([]card.Card, 0)
		p.Bet = 0
		p.TotalContributed = 0
		p.Folded = false
		p.AllIn = false
		p.Acted = false
		p.HandStrength = ""
		p.Reaction = ""
		p.ExposedCards = false
	}
}

func (g *Game) AddPlayer(id, name string, chips int, seat int) error {
	if len(g.Players) >= 8 {
		return errors.New("table full (max 8 players)")
	}
	if g.Phase != PhaseWaiting {
		return errors.New("the game is already underway")
	}
	if chips <= 0 {
		return errors.New("you have no chips — please rebuy before joining")
	}
	for _, p := range g.Players {
		if p.ID == id {
			return errors.New("you are already seated at this table")
		}
		if strings.EqualFold(p.Name, name) {
			return fmt.Errorf("player name '%s' is already taken", name)
		}
		if p.Seat == seat && seat >= 0 && seat <= 7 {
			return fmt.Errorf("seat %d is already occupied", seat)
		}
	}
	if seat < 0 || seat > 7 {
		occupied := make(map[int]bool)
		for _, p := range g.Players {
			occupied[p.Seat] = true
		}
		seat = -1
		for i := 0; i < 8; i++ {
			if !occupied[i] {
				seat = i
				break
			}
		}
		if seat == -1 {
			return errors.New("no seats available")
		}
	}
	g.Players = append(g.Players, &Player{
		ID:              id,
		Name:            name,
		Chips:           chips,
		Seat:            seat,
		RebuysRemaining: g.MaxRebuys,
	})
	return nil
}

func (g *Game) NextBlindLevel() (nextSB, nextBB int) {
	newSB := g.SmallBlind * 2
	newBB := g.BigBlind * 2
	maxStack := 0
	for _, p := range g.Players {
		if !p.SittingOut && p.Chips > maxStack {
			maxStack = p.Chips
		}
	}
	if maxStack > 0 && newBB > maxStack/4 {

		return g.SmallBlind, g.BigBlind
	}
	return newSB, newBB
}

func (g *Game) Start() error {
	activeCount := 0
	for _, p := range g.Players {
		if p.Chips > 0 && !p.SittingOut {
			activeCount++
		}
	}
	if activeCount < 2 {
		return errors.New("requires at least 2 players (not sitting out) with chips to start a deal")
	}

	if g.BlindEscalationMins > 0 {
		if g.NextRaiseTime.IsZero() {
			g.NextRaiseTime = time.Now().Add(time.Duration(g.BlindEscalationMins) * time.Minute)
		} else if time.Now().After(g.NextRaiseTime) {
			g.SmallBlind, g.BigBlind = g.NextBlindLevel()
			g.NextRaiseTime = time.Now().Add(time.Duration(g.BlindEscalationMins) * time.Minute)
		}
	} else {
		g.NextRaiseTime = time.Time{}
	}

	g.HandCount++

	g.Deck = deck.New()

	g.Board = make([]card.Card, 0)
	g.Pot = 0
	g.CurrentBet = 0
	g.LastWinners = nil

	for _, p := range g.Players {
		p.Hole = make([]card.Card, 0, 2)
		p.Bet = 0
		p.TotalContributed = 0
		p.AllIn = false
		p.Acted = false
		p.HandStrength = ""
		p.TimeoutCount = 0
		p.IsVPIPThisHand = false
		p.ExposedCards = false
		if p.Chips <= 0 || p.SittingOut {
			p.Folded = true
		} else {
			p.Folded = false
			p.HandsPlayed++
		}
	}

	for _, p := range g.Players {
		if !p.Folded {
			cards := g.Deck.Deal(2)
			p.Hole = append([]card.Card(nil), cards...)
		}
	}

	if g.HandCount == 1 {
		g.DealerIdx = rand.Intn(len(g.Players)) - 1
		if g.DealerIdx < 0 {
			g.DealerIdx = len(g.Players) - 1
		}
	}

	dealerIdx := g.nextActivePlayerIdx(g.DealerIdx)
	if dealerIdx < 0 {
		return errors.New("no active players found for dealer assignment")
	}
	g.DealerIdx = dealerIdx

	var sbIdx, bbIdx int
	if activeCount == 2 {
		sbIdx = g.DealerIdx
		bbIdx = g.nextActivePlayerIdx(g.DealerIdx)
	} else {
		sbIdx = g.nextActivePlayerIdx(g.DealerIdx)
		if sbIdx < 0 {
			return errors.New("no active player for small blind")
		}
		bbIdx = g.nextActivePlayerIdx(sbIdx)
		if bbIdx < 0 {
			return errors.New("no active player for big blind")
		}
	}

	g.SmallBlindIdx = sbIdx
	g.BigBlindIdx = bbIdx
	g.LastRaiseSize = g.BigBlind

	g.postBlind(g.Players[sbIdx], g.SmallBlind)
	g.postBlind(g.Players[bbIdx], g.BigBlind)

	g.CurrentBet = g.BigBlind
	g.Phase = PhasePreFlop

	actorsCount := 0
	for _, p := range g.Players {
		if !p.Folded && !p.AllIn {
			actorsCount++
		}
	}

	if actorsCount <= 1 {
		for g.Phase != PhaseWaiting {
			g.advancePhase()
		}
	} else {
		g.ActiveIdx = (bbIdx + 1) % len(g.Players)
		for g.Players[g.ActiveIdx].Folded || g.Players[g.ActiveIdx].AllIn {
			g.ActiveIdx = (g.ActiveIdx + 1) % len(g.Players)
		}
		g.ActivePlayerID = g.Players[g.ActiveIdx].ID
	}

	return nil
}

func (g *Game) playerContribute(p *Player, amount int) {
	if amount > p.Chips {
		amount = p.Chips
	}
	p.Chips -= amount
	p.Bet += amount
	p.TotalContributed += amount
	g.Pot += amount
	if p.Chips == 0 {
		p.AllIn = true
	}
}

func (g *Game) nextActivePlayerIdx(startIdx int) int {
	if len(g.Players) == 0 {
		return -1
	}
	idx := (startIdx + 1) % len(g.Players)
	for i := 0; i < len(g.Players); i++ {
		if !g.Players[idx].Folded {
			return idx
		}
		idx = (idx + 1) % len(g.Players)
	}
	return -1
}

func (g *Game) postBlind(p *Player, amount int) {
	g.playerContribute(p, amount)
}

func (g *Game) Act(playerID string, action Action, amount int) error {
	if g.Phase == PhaseShowdown || g.Phase == PhaseWaiting {
		return errors.New("game is not in active phase")
	}

	player := g.Players[g.ActiveIdx]

	if player.ID != playerID {
		return fmt.Errorf("not your turn, it's %s turn", player.Name)
	}

	player.TimeoutCount = 0

	if player.Folded || player.AllIn {
		return errors.New("player cannot act")
	}

	if player.Chips <= 0 {
		return errors.New("player has no chips and cannot act")
	}

	switch action {
	case ActionCall, ActionRaise, ActionAllIn:
		if g.Phase == PhasePreFlop {
			if !player.IsVPIPThisHand {
				player.IsVPIPThisHand = true
				player.HandsVPIP++
			}
		}
	}

	switch action {

	case ActionFold:
		player.Folded = true

	case ActionCheck:
		if player.Bet != g.CurrentBet {
			return fmt.Errorf("cannot check, need to call %d", g.CurrentBet-player.Bet)
		}

	case ActionCall:
		toCall := g.CurrentBet - player.Bet
		if toCall <= 0 {
			return errors.New("nothing to call, use check")
		}
		g.playerContribute(player, toCall)

	case ActionRaise:
		minRaise := g.CurrentBet + g.LastRaiseSize
		if amount < minRaise {
			if amount-player.Bet != player.Chips {
				return fmt.Errorf("raise must be to at least %d", minRaise)
			}
		}

		diff := amount - player.Bet
		if diff > player.Chips {
			return errors.New("not enough chips")
		}

		raiseSize := amount - g.CurrentBet
		g.playerContribute(player, diff)
		g.CurrentBet = amount
		if raiseSize > g.LastRaiseSize {
			g.LastRaiseSize = raiseSize
		}
		g.resetOtherPlayersActed(player)

	case ActionAllIn:
		g.playerContribute(player, player.Chips)

		if player.Bet > g.CurrentBet {
			raiseSize := player.Bet - g.CurrentBet
			g.CurrentBet = player.Bet
			if raiseSize > g.LastRaiseSize {
				g.LastRaiseSize = raiseSize
			}
			g.resetOtherPlayersActed(player)
		}
	}

	player.Acted = true

	if g.countNonFolded() == 1 {
		g.Showdown()
		return nil
	}

	if g.isRoundOver() {
		g.advancePhase()
		actorsCount := 0
		for _, p := range g.Players {
			if !p.Folded && !p.AllIn {
				actorsCount++
			}
		}
		if actorsCount <= 1 {
			for g.Phase != PhaseWaiting {
				g.advancePhase()
			}
		}
		return nil
	}

	g.nextActivePlayer()
	g.ActivePlayerID = g.Players[g.ActiveIdx].ID
	return nil
}

func (g *Game) isRoundOver() bool {
	for _, p := range g.Players {
		if !p.Folded && !p.AllIn {
			if !p.Acted || p.Bet < g.CurrentBet {
				return false
			}
		}
	}
	return true
}

func (g *Game) resetOtherPlayersActed(activePlayer *Player) {
	for _, p := range g.Players {
		if p != activePlayer && !p.Folded && !p.AllIn {
			p.Acted = false
		}
	}
}

func (g *Game) countNonFolded() int {
	n := 0
	for _, p := range g.Players {
		if !p.Folded {
			n++
		}
	}
	return n
}

func (g *Game) nextPhase() {
	for _, p := range g.Players {
		p.Bet = 0
		p.Acted = false
	}
	g.CurrentBet = 0
	g.LastRaiseSize = g.BigBlind

	g.ActiveIdx = (g.DealerIdx + 1) % len(g.Players)
	g.skipFolded()
	if g.Phase != PhaseWaiting && len(g.Players) > 0 {
		g.ActivePlayerID = g.Players[g.ActiveIdx].ID
	}

	switch g.Phase {
	case PhasePreFlop:
		g.Phase = PhaseFlop

	case PhaseFlop:
		g.Phase = PhaseTurn

	case PhaseTurn:
		g.Phase = PhaseRiver

	case PhaseRiver:
		g.Phase = PhaseShowdown
		g.Showdown()
	default:
		return
	}
}

func (g *Game) dealBoard() {
	switch g.Phase {
	case PhaseWaiting, PhaseShowdown:

	case PhasePreFlop:

	case PhaseFlop:
		g.Board = append(g.Board, g.Deck.Deal(3)...)

	case PhaseTurn:
		g.Board = append(g.Board, g.Deck.Deal(1)...)

	case PhaseRiver:
		g.Board = append(g.Board, g.Deck.Deal(1)...)
	default:
		return
	}
}

func (g *Game) advancePhase() {
	g.nextPhase()
	g.dealBoard()
}

func (g *Game) nextActivePlayer() {
	g.ActiveIdx = (g.ActiveIdx + 1) % len(g.Players)
	g.skipFolded()
}

func (g *Game) skipFolded() {
	hasActors := false
	for _, p := range g.Players {
		if !p.Folded && !p.AllIn {
			hasActors = true
			break
		}
	}
	if !hasActors {
		return
	}

	for g.Players[g.ActiveIdx].Folded || g.Players[g.ActiveIdx].AllIn {
		g.ActiveIdx = (g.ActiveIdx + 1) % len(g.Players)
	}
}

func compareTieBreak(a, b []int) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] > b[i] {
			return 1
		}
		if a[i] < b[i] {
			return -1
		}
	}
	return 0
}

type evaluated struct {
	player *Player
	result hand.Result
}

type SubPot struct {
	Amount       int      `json:"amount"`
	Contributors []string `json:"contributors"`
}

func (g *Game) CalculateCurrentSubPots() []SubPot {
	if g.Phase == PhaseWaiting || g.Phase == PhaseShowdown {
		return nil
	}

	type tempPlayer struct {
		name             string
		totalContributed int
		folded           bool
	}

	var temps []tempPlayer
	for _, p := range g.Players {
		if p.TotalContributed > 0 {
			temps = append(temps, tempPlayer{
				name:             p.Name,
				totalContributed: p.TotalContributed,
				folded:           p.Folded,
			})
		}
	}

	var subPots []SubPot
	for {
		minBet := 0
		for _, tp := range temps {
			if !tp.folded && tp.totalContributed > 0 {
				if minBet == 0 || tp.totalContributed < minBet {
					minBet = tp.totalContributed
				}
			}
		}

		if minBet == 0 {
			break
		}

		subPotAmt := 0
		var contributors []string
		for i := range temps {
			tp := &temps[i]
			if tp.totalContributed > 0 {
				take := minBet
				if tp.totalContributed < take {
					take = tp.totalContributed
				}
				subPotAmt += take
				tp.totalContributed -= take

				if !tp.folded {
					contributors = append(contributors, tp.name)
				}
			}
		}

		if subPotAmt > 0 && len(contributors) > 0 {
			subPots = append(subPots, SubPot{
				Amount:       subPotAmt,
				Contributors: contributors,
			})
		}
	}

	return subPots
}

func (g *Game) getNextSubPot() (int, []*Player) {
	minBet := 0
	for _, p := range g.Players {
		if !p.Folded && p.TotalContributed > 0 {
			if minBet == 0 || p.TotalContributed < minBet {
				minBet = p.TotalContributed
			}
		}
	}

	if minBet == 0 {
		return 0, nil
	}

	subPot := 0
	var contributors []*Player
	for _, p := range g.Players {
		if p.TotalContributed > 0 {
			take := minBet
			if p.TotalContributed < take {
				take = p.TotalContributed
			}
			subPot += take
			p.TotalContributed -= take

			if !p.Folded {
				contributors = append(contributors, p)
			}
		}
	}
	return subPot, contributors
}

func (g *Game) evaluateHands(contributors []*Player, playerHands map[string]hand.Result) []evaluated {
	var evals []evaluated
	for _, p := range contributors {
		all := append(p.Hole, g.Board...)
		r := hand.Evaluate(all)
		evals = append(evals, evaluated{p, r})
		playerHands[p.ID] = r
	}
	return evals
}

func findWinners(evals []evaluated) []*Player {
	if len(evals) == 0 {
		return nil
	}
	best := evals[0].result
	for _, e := range evals[1:] {
		if e.result.Rank > best.Rank {
			best = e.result
		} else if e.result.Rank == best.Rank && compareTieBreak(e.result.TieBreak, best.TieBreak) > 0 {
			best = e.result
		}
	}

	var subWinners []*Player
	for _, e := range evals {
		if e.result.Rank == best.Rank && compareTieBreak(e.result.TieBreak, best.TieBreak) == 0 {
			subWinners = append(subWinners, e.player)
		}
	}
	return subWinners
}

func (g *Game) Showdown() []Winner {
	g.Phase = PhaseWaiting

	var active []*Player
	for _, p := range g.Players {
		if !p.Folded {
			active = append(active, p)
		}
	}

	if len(active) == 1 {
		won := g.Pot
		active[0].Chips += won
		g.Pot = 0

		for _, p := range g.Players {
			if p.Chips <= 0 {
				p.SittingOut = true
			}
		}

		if won > active[0].BiggestPotWon {
			active[0].BiggestPotWon = won
		}

		winners := []Winner{{Player: active[0], Amount: won}}
		g.LastWinners = winners
		return winners
	}

	winnings := make(map[string]int)
	playerHands := make(map[string]hand.Result)

	for {
		subPot, contributors := g.getNextSubPot()
		if subPot == 0 || len(contributors) == 0 {
			break
		}

		evals := g.evaluateHands(contributors, playerHands)
		subWinners := findWinners(evals)

		share := subPot / len(subWinners)
		remainder := subPot % len(subWinners)
		for _, w := range subWinners {
			winnings[w.ID] += share
		}
		if remainder > 0 {
			for _, p := range g.Players {
				if _, ok := findPlayerInSlice(subWinners, p.ID); ok {
					winnings[p.ID] += remainder
					break
				}
			}
		}
	}

	var winners []Winner
	for _, p := range g.Players {
		amt := winnings[p.ID]
		if amt > 0 {
			p.Chips += amt
			h, ok := playerHands[p.ID]
			if !ok {
				all := append(p.Hole, g.Board...)
				h = hand.Evaluate(all)
			}
			if amt > p.BiggestPotWon {
				p.BiggestPotWon = amt
			}
			winners = append(winners, Winner{
				Player: p,
				Amount: amt,
				Hand:   h,
			})
		}
	}
	g.Pot = 0

	for _, p := range g.Players {
		if p.Chips <= 0 {
			p.SittingOut = true
		}
	}

	g.LastWinners = winners
	return winners
}

type Winner struct {
	Player *Player
	Hand   hand.Result
	Amount int
}

func (g *Game) Status() string {
	result := fmt.Sprintf("=== Phase: %s | Bank: %d ===\n", g.Phase, g.Pot)
	result += fmt.Sprintf("Board: %v\n", g.Board)
	result += "Players:\n"
	for i, p := range g.Players {
		marker := "  "
		if i == g.ActiveIdx && g.Phase != PhaseWaiting && g.Phase != PhaseShowdown {
			marker = "→ "
		}
		status := ""
		if p.Folded {
			status = " [FOLD]"
		} else if p.AllIn {
			status = " [ALL-IN]"
		}
		result += fmt.Sprintf("%s%s | Chips: %d | Bet: %d%s\n",
			marker, p.Name, p.Chips, p.Bet, status)
	}
	return result
}

func (g *Game) SitOut(playerID string) error {
	for i, p := range g.Players {
		if p.ID == playerID {
			if p.SittingOut {
				return nil
			}
			p.SittingOut = true
			p.TimeoutCount = 0
			if g.Phase != PhaseWaiting && g.Phase != PhaseShowdown {
				if g.ActiveIdx == i && !p.Folded && !p.AllIn {
					action := ActionFold
					if p.Bet == g.CurrentBet {
						action = ActionCheck
					}
					if actErr := g.Act(playerID, action, 0); actErr != nil {
						slog.Error("SitOut: Act failed", "playerID", playerID, "action", action, "err", actErr)
					}
				} else {
					p.Folded = true
					if g.countNonFolded() == 1 {
						g.Showdown()
					}
				}
			}
			return nil
		}
	}
	return errors.New("player not found")
}

func (g *Game) SitIn(playerID string) error {
	for _, p := range g.Players {
		if p.ID == playerID {
			p.SittingOut = false
			p.TimeoutCount = 0
			return nil
		}
	}
	return errors.New("player not found")
}

func (g *Game) RemovePlayer(playerID string) error {
	idx := -1
	for i, p := range g.Players {
		if p.ID == playerID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("player not found")
	}

	if g.Phase != PhaseWaiting && g.Phase != PhaseShowdown {
		p := g.Players[idx]
		if !p.Folded {
			if g.ActiveIdx == idx {
				action := ActionFold
				if p.Bet == g.CurrentBet {
					action = ActionCheck
				}
				_ = g.Act(playerID, action, 0)
			} else {
				p.Folded = true
				if g.countNonFolded() == 1 {
					g.Showdown()
				}
			}
		}
	}

	activeID := g.ActivePlayerID
	var dealerID string
	if g.DealerIdx >= 0 && g.DealerIdx < len(g.Players) {
		dealerID = g.Players[g.DealerIdx].ID
	}

	g.Players = append(g.Players[:idx], g.Players[idx+1:]...)

	g.ActiveIdx = -1
	g.DealerIdx = -1
	for i, p := range g.Players {
		if p.ID == activeID {
			g.ActiveIdx = i
		}
		if p.ID == dealerID {
			g.DealerIdx = i
		}
	}

	if len(g.Players) == 0 {
		g.Phase = PhaseWaiting
		g.ActiveIdx = -1
		g.DealerIdx = -1
		g.ActivePlayerID = ""
	} else {
		if g.DealerIdx == -1 {
			g.DealerIdx = 0
		}
		if g.Phase != PhaseWaiting {
			if g.ActiveIdx == -1 {
				g.ActiveIdx = idx % len(g.Players)
				g.skipFolded()
			}
			if g.ActiveIdx >= 0 && g.ActiveIdx < len(g.Players) {
				g.ActivePlayerID = g.Players[g.ActiveIdx].ID
			}
		} else {
			g.ActiveIdx = -1
			g.ActivePlayerID = ""
		}
	}

	return nil
}

func findPlayerInSlice(players []*Player, playerID string) (*Player, bool) {
	for _, p := range players {
		if p.ID == playerID {
			return p, true
		}
	}
	return nil, false
}

func (g *Game) ExposeCards(playerID string, show bool) error {
	for _, p := range g.Players {
		if p.ID == playerID {
			if len(p.Hole) == 0 {
				return errors.New("no cards to show")
			}
			p.ExposedCards = show
			return nil
		}
	}
	return errors.New("player not found")
}
