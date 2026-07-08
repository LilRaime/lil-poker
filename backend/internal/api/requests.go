package api

type CreateRoomRequest struct {
	Name                string `json:"name"`
	SmallBlind          int    `json:"small_blind"`
	BigBlind            int    `json:"big_blind"`
	MaxPlayers          int    `json:"max_players"`
	BlindEscalationMins *int   `json:"blind_escalation_mins,omitempty"`
	StartingChips       *int   `json:"starting_chips,omitempty"`
	MaxRebuys           int    `json:"max_rebuys,omitempty"`
	TurnTimeoutSecs     int    `json:"turn_timeout_secs,omitempty"`
}

type CreateGameRequest struct {
	SmallBlind int `json:"small_blind"`
	BigBlind   int `json:"big_blind"`
}

type AddPlayerRequest struct {
	UUID string `json:"uuid"`
	Seat *int   `json:"seat,omitempty"`
}

type ActRequest struct {
	PlayerID string `json:"player_id"`
	Action   string `json:"action"`
	Amount   int    `json:"amount,omitempty"`
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RebuyRequest struct {
	UUID string `json:"uuid"`
}

type SitRequest struct {
	Action string `json:"action"`
}

type UpdateBlindsRequest struct {
	SmallBlind int `json:"small_blind"`
	BigBlind   int `json:"big_blind"`
}
