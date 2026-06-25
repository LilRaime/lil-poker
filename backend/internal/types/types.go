package types

import "lil-poker/internal/card"

type ChatMessage struct {
	PlayerName string `json:"player_name"`
	PlayerID   string `json:"player_id,omitempty"`
	Text       string `json:"text"`
	Time       int64  `json:"time"`
	System     bool   `json:"system,omitempty"`
}

type WSMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type PlayerStatus struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Chips         int         `json:"chips"`
	Hole          []card.Card `json:"hole,omitempty"`
	Bet           int         `json:"bet"`
	Folded        bool        `json:"folded"`
	AllIn         bool        `json:"all_in"`
	Acted         bool        `json:"acted"`
	CurrentHand   string      `json:"current_hand,omitempty"`
	SittingOut    bool        `json:"sitting_out"`
	Seat          int         `json:"seat"`
	Reaction      string      `json:"reaction,omitempty"`
	HandsPlayed   int         `json:"hands_played"`
	HandsVPIP     int         `json:"hands_vpip"`
	BiggestPotWon int         `json:"biggest_pot_won"`
	IsSmallBlind  bool        `json:"is_small_blind"`
	IsBigBlind    bool        `json:"is_big_blind"`
}

type WinnerStatus struct {
	PlayerID   string      `json:"player_id"`
	PlayerName string      `json:"player_name"`
	HandRank   string      `json:"hand_rank,omitempty"`
	HandCards  []card.Card `json:"hand_cards,omitempty"`
	Amount     int         `json:"amount"`
}

type HandHistoryEntry struct {
	HandNum int            `json:"hand_num"`
	Board   []card.Card    `json:"board"`
	Winners []WinnerStatus `json:"winners"`
}

type SubPot struct {
	Amount       int      `json:"amount"`
	Contributors []string `json:"contributors"`
}

type GameStateResponse struct {
	Players             []PlayerStatus     `json:"players"`
	Board               []card.Card        `json:"board"`
	Phase               string             `json:"phase"`
	Pot                 int                `json:"pot"`
	CurrentBet          int                `json:"current_bet"`
	ActiveIdx           int                `json:"active_idx"`
	ActiveName          string             `json:"active_name,omitempty"`
	ActivePlayerID      string             `json:"active_player_id,omitempty"`
	DealerIdx           int                `json:"dealer_idx"`
	SmallBlind          int                `json:"small_blind"`
	BigBlind            int                `json:"big_blind"`
	LastWinners         []WinnerStatus     `json:"last_winners,omitempty"`
	ActionDeadline      int64              `json:"action_deadline"`
	HandCount           int                `json:"hand_count"`
	ChatMessages        []ChatMessage      `json:"chat_messages,omitempty"`
	NextSmallBlind      int                `json:"next_small_blind"`
	NextBigBlind        int                `json:"next_big_blind"`
	HandsUntilRaise     int                `json:"hands_until_raise"`
	ObserverCount       int                `json:"observer_count"`
	BlindsRaiseDeadline int64              `json:"blinds_raise_deadline"`
	HandHistory         []HandHistoryEntry `json:"hand_history,omitempty"`
	CreatorID           string             `json:"creator_id,omitempty"`
	SubPots             []SubPot           `json:"sub_pots,omitempty"`
	Observers           []string           `json:"observers,omitempty"`
}
