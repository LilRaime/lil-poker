package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"lil-poker/internal/game"
	"lil-poker/internal/room"
	"lil-poker/internal/store"
	"lil-poker/internal/types"
)

func (s *Server) handleListRooms(w http.ResponseWriter, r *http.Request) {
	rooms := s.rm.ListRooms()
	if rooms == nil {
		rooms = []room.RoomInfo{}
	}
	writeJSON(w, http.StatusOK, rooms)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	rooms := s.rm.ListRooms()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"rooms":  len(rooms),
	})
}

func (s *Server) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	s.RequireAuth(func(w http.ResponseWriter, r *http.Request, user *store.User) {
		defaultName := "New Table"
		var creatorID string
		if user != nil {
			defaultName = user.Username + "'s Table"
			creatorID = user.UUID
		}

		var req CreateRoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.ContentLength > 0 {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" {
			req.Name = defaultName
		}
		if req.SmallBlind <= 0 {
			req.SmallBlind = 10
		}
		if req.BigBlind <= 0 {
			req.BigBlind = 20
		}
		if req.BigBlind < req.SmallBlind {
			writeError(w, http.StatusBadRequest, "big blind must be greater than or equal to small blind")
			return
		}
		if req.MaxPlayers < 2 || req.MaxPlayers > 8 {
			req.MaxPlayers = 8
		}
		escalationMins := 5
		if req.BlindEscalationMins != nil {
			escalationMins = *req.BlindEscalationMins
			if escalationMins < 0 {
				escalationMins = 0
			}
		}

		startingChips := 1000
		if req.StartingChips != nil {
			startingChips = *req.StartingChips
			if startingChips < 0 {
				startingChips = 0
			}
		}
		maxRebuys := req.MaxRebuys
		if maxRebuys <= 0 {
			maxRebuys = 3
		}
		turnTimeout := req.TurnTimeoutSecs
		if turnTimeout < 5 || turnTimeout > 300 {
			turnTimeout = 20
		}

		roomID := r.URL.Query().Get("room")
		r2 := s.rm.CreateRoom(roomID, req.Name, creatorID, req.MaxPlayers, req.SmallBlind, req.BigBlind, escalationMins, startingChips, maxRebuys, turnTimeout)
		writeJSON(w, http.StatusCreated, r2.Info())
	})(w, r)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	r2 := s.roomFromRequest(w, r)
	if r2 == nil {
		return
	}
	s.RequireAuth(func(w http.ResponseWriter, r *http.Request, _ *store.User) {
		var req CreateGameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.ContentLength > 0 {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.SmallBlind <= 0 {
			req.SmallBlind = 10
		}
		if req.BigBlind <= 0 {
			req.BigBlind = 20
		}
		if req.BigBlind < req.SmallBlind {
			writeError(w, http.StatusBadRequest, "big blind must be greater than or equal to small blind")
			return
		}
		r2.Sg.Lock()
		r2.Sg.GetGame().Reset(req.SmallBlind, req.BigBlind)
		r2.Sg.SetActionDeadline(0)
		r2.Sg.Unlock()

		writeJSON(w, http.StatusOK, map[string]string{"message": "game created successfully"})
		s.broadcast(r2)
	})(w, r)
}

func (s *Server) handleAddPlayer(w http.ResponseWriter, r *http.Request) {
	r2 := s.roomFromRequest(w, r)
	if r2 == nil {
		return
	}

	req := AddPlayerRequest{}
	_ = json.NewDecoder(r.Body).Decode(&req)

	seatVal := -1
	if req.Seat != nil {
		seatVal = *req.Seat
	}

	user, err := s.getAuthenticatedUser(r)
	if err != nil {
		slog.Error("getAuthenticatedUser failed", "handler", "handleAddPlayer", "err", err)
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	if user == nil && req.UUID != "" && s.bypassAuth {
		user, err = store.GetUserByUUID(s.db, req.UUID)
		if err != nil {
			slog.Error("GetUserByUUID failed", "handler", "handleAddPlayer", "uuid", req.UUID, "err", err)
			writeError(w, http.StatusInternalServerError, "database error")
			return
		}
	}

	if user == nil {
		writeError(w, http.StatusUnauthorized, "invalid session user")
		return
	}

	r2.Sg.Lock()
	if len(r2.Sg.GetGame().Players) >= r2.MaxPlayers {
		r2.Sg.Unlock()
		writeError(w, http.StatusBadRequest, fmt.Sprintf("room is full (max %d players)", r2.MaxPlayers))
		return
	}
	playerChips := user.Chips
	if r2.StartingChips > 0 {
		playerChips = r2.StartingChips
	}
	err = r2.Sg.GetGame().AddPlayer(user.UUID, user.Username, playerChips, seatVal)
	r2.Sg.Unlock()

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if r2.StartingChips == 0 {
		if errDb := store.ResetUserRebuys(s.db, user.UUID, 3); errDb != nil {
			slog.Error("failed to reset player rebuys on table join", "uuid", user.UUID, "err", errDb)
		}
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":    user.UUID,
		"name":  user.Username,
		"chips": playerChips,
	})
	s.broadcast(r2)
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	r2 := s.roomFromRequest(w, r)
	if r2 == nil {
		return
	}
	s.RequireAuth(func(w http.ResponseWriter, r *http.Request, user *store.User) {
		if user != nil && r2.CreatorID != "" && user.UUID != r2.CreatorID {
			writeError(w, http.StatusForbidden, "only the room creator can start the game")
			return
		}

		var startMsg string
		var chipUpdates map[string]int
		r2.Sg.Lock()
		g := r2.Sg.GetGame()
		err := g.Start()
		if err == nil {
			r2.Sg.UpdateHandEvaluations()
			startMsg = fmt.Sprintf("Hand #%d started. Blinds are %d/%d.", g.HandCount, g.SmallBlind, g.BigBlind)
			if g.Phase == game.PhaseWaiting {
				chipUpdates = make(map[string]int, len(g.Players))
				for _, p := range g.Players {
					chipUpdates[p.ID] = p.Chips
				}
			}
		}
		r2.Sg.Unlock()

		var status types.GameStateResponse
		if err == nil {
			status = r2.Sg.GetStatus("", r2.Wsm.GetObserverCount(), r2.StartingChips)
		}

		if r2.StartingChips == 0 {
			s.flushChipsAsync(chipUpdates)
		}

		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		if startMsg != "" {
			r2.Sg.AddSystemMessage(startMsg)
		}

		writeJSON(w, http.StatusOK, status)
		s.broadcast(r2)
		s.resetTimer(r2)
	})(w, r)
}

func (s *Server) handleAct(w http.ResponseWriter, r *http.Request) {
	r2 := s.roomFromRequest(w, r)
	if r2 == nil {
		return
	}

	var req ActRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := s.getAuthenticatedUser(r)
	if err != nil {
		slog.Error("getAuthenticatedUser failed", "handler", "handleAct", "err", err)
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	var playerID string
	if user != nil {
		playerID = user.UUID
	} else if s.bypassAuth {
		playerID = req.PlayerID
	}

	if playerID == "" {
		writeError(w, http.StatusUnauthorized, "must be logged in")
		return
	}

	action, err := parseAction(req.Action)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var actMsg string
	var showdownWinners []types.WinnerStatus
	var chipUpdates map[string]int
	r2.Sg.Lock()
	var playerName string
	var toCall int
	g := r2.Sg.GetGame()
	for _, p := range g.Players {
		if p.ID == playerID {
			playerName = p.Name
			toCall = g.CurrentBet - p.Bet
			break
		}
	}

	err = g.Act(playerID, action, req.Amount)
	if err == nil {
		r2.Sg.UpdateHandEvaluations()
		r2.Sg.CheckRecordHandHistoryLocked()

		switch action {
		case game.ActionFold:
			actMsg = fmt.Sprintf("%s folded.", playerName)
		case game.ActionCheck:
			actMsg = fmt.Sprintf("%s checked.", playerName)
		case game.ActionCall:
			actMsg = fmt.Sprintf("%s called %d.", playerName, toCall)
		case game.ActionRaise:
			actMsg = fmt.Sprintf("%s raised to %d.", playerName, req.Amount)
		case game.ActionAllIn:
			actMsg = fmt.Sprintf("%s went All-In.", playerName)
		}

		if g.Phase == game.PhaseWaiting {
			chipUpdates = make(map[string]int, len(g.Players))
			for _, p := range g.Players {
				chipUpdates[p.ID] = p.Chips
			}
			if len(g.LastWinners) > 0 {
				showdownWinners = make([]types.WinnerStatus, len(g.LastWinners))
				for i, w := range g.LastWinners {
					ws := types.WinnerStatus{
						PlayerName: w.Player.Name,
						Amount:     w.Amount,
					}
					if len(w.Hand.Cards) > 0 {
						ws.HandRank = w.Hand.Description()
					}
					showdownWinners[i] = ws
				}
			}
		}
	}
	r2.Sg.Unlock()

	var status types.GameStateResponse
	if err == nil {
		status = r2.Sg.GetStatus(playerID, r2.Wsm.GetObserverCount(), r2.StartingChips)
	}

	if r2.StartingChips == 0 {
		s.flushChipsAsync(chipUpdates)
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if actMsg != "" {
		r2.Sg.AddSystemMessage(actMsg)
	}
	for _, w := range showdownWinners {
		desc := ""
		if w.HandRank != "" {
			desc = fmt.Sprintf(" with %s", w.HandRank)
		}
		r2.Sg.AddSystemMessage(fmt.Sprintf("%s won %d chips%s.", w.PlayerName, w.Amount, desc))
	}

	writeJSON(w, http.StatusOK, status)
	s.broadcast(r2)
	s.resetTimer(r2)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	r2 := s.roomFromRequest(w, r)
	if r2 == nil {
		return
	}

	user, _ := s.getAuthenticatedUser(r)
	playerID := ""
	if user != nil {
		playerID = user.UUID
	}

	writeJSON(w, http.StatusOK, r2.Sg.GetStatus(playerID, r2.Wsm.GetObserverCount(), r2.StartingChips))
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	r2 := s.roomFromRequest(w, r)
	if r2 == nil {
		return
	}

	user, err := s.getAuthenticatedUser(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if user == nil && !s.bypassAuth {
		writeError(w, http.StatusUnauthorized, "session invalid")
		return
	}

	playerID := ""
	username := "observer"
	if user != nil {
		playerID = user.UUID
		username = user.Username
	}

	r2.Wsm.DisconnectPlayer(playerID)

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn.SetReadLimit(512)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	client := room.NewWsClient(conn, playerID, username)
	s.registerClient(r2, client)

	go s.writePump(r2, client)
	go s.readPump(r2, client)

	initialState := r2.Sg.GetStatus(playerID, r2.Wsm.GetObserverCount(), r2.StartingChips)
	initialState.Observers = r2.Wsm.GetObserverNames()
	client.Send <- initialState
}

func (s *Server) handleSit(w http.ResponseWriter, r *http.Request) {
	r2 := s.roomFromRequest(w, r)
	if r2 == nil {
		return
	}

	var req SitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := s.getAuthenticatedUser(r)
	if err != nil {
		slog.Error("getAuthenticatedUser failed", "handler", "handleSit", "err", err)
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	r2.Sg.Lock()
	var sitErr error
	g := r2.Sg.GetGame()
	switch req.Action {
	case "in":
		sitErr = g.SitIn(user.UUID)
	case "out":
		sitErr = g.SitOut(user.UUID)
	default:
		r2.Sg.Unlock()
		writeError(w, http.StatusBadRequest, "invalid sit action (must be 'in' or 'out')")
		return
	}

	if sitErr != nil {
		r2.Sg.Unlock()
		writeError(w, http.StatusBadRequest, sitErr.Error())
		return
	}

	r2.Sg.UpdateHandEvaluations()
	r2.Sg.Unlock()

	var msg string
	if req.Action == "in" {
		msg = fmt.Sprintf("%s sat back in.", user.Username)
	} else {
		msg = fmt.Sprintf("%s is sitting out.", user.Username)
	}
	r2.Sg.AddSystemMessage(msg)

	writeJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("successfully sat %s", req.Action)})
	s.broadcast(r2)
	s.resetTimer(r2)
}

func (s *Server) handleStand(w http.ResponseWriter, r *http.Request) {
	r2 := s.roomFromRequest(w, r)
	if r2 == nil {
		return
	}

	user, err := s.getAuthenticatedUser(r)
	if err != nil {
		slog.Error("getAuthenticatedUser failed", "handler", "handleStand", "err", err)
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var finalChips int
	r2.Sg.Lock()
	g := r2.Sg.GetGame()
	
	found := false
	for _, p := range g.Players {
		if p.ID == user.UUID {
			finalChips = p.Chips + p.Bet
			found = true
			break
		}
	}

	if !found {
		r2.Sg.Unlock()
		writeError(w, http.StatusBadRequest, "player is not seated at this table")
		return
	}

	standErr := g.RemovePlayer(user.UUID)
	r2.Sg.UpdateHandEvaluations()
	r2.Sg.CheckRecordHandHistoryLocked()
	r2.Sg.Unlock()

	if standErr != nil {
		writeError(w, http.StatusBadRequest, standErr.Error())
		return
	}

	if r2.StartingChips == 0 {
		if errDb := s.updateUserChips(user.UUID, finalChips); errDb != nil {
			slog.Error("failed to update user chips on stand up", "uuid", user.UUID, "err", errDb)
		}
	}

	r2.Sg.AddSystemMessage(fmt.Sprintf("%s stood up and left the table.", user.Username))

	writeJSON(w, http.StatusOK, map[string]string{"message": "successfully stood up"})
	s.broadcast(r2)
	s.resetTimer(r2)
}

func (s *Server) handleUpdateBlinds(w http.ResponseWriter, r *http.Request) {
	r2 := s.roomFromRequest(w, r)
	if r2 == nil {
		return
	}
	s.RequireAuth(func(w http.ResponseWriter, r *http.Request, _ *store.User) {
		var req UpdateBlindsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.SmallBlind <= 0 || req.BigBlind <= 0 {
			writeError(w, http.StatusBadRequest, "blinds must be greater than zero")
			return
		}
		if req.BigBlind < req.SmallBlind {
			writeError(w, http.StatusBadRequest, "big blind must be greater than or equal to small blind")
			return
		}

		r2.Sg.Lock()
		g := r2.Sg.GetGame()
		g.SmallBlind = req.SmallBlind
		g.BigBlind = req.BigBlind
		r2.Sg.Unlock()

		writeJSON(w, http.StatusOK, map[string]string{"message": "blinds updated successfully"})
		s.broadcast(r2)
	})(w, r)
}

func (s *Server) handleShowCards(w http.ResponseWriter, r *http.Request) {
	r2 := s.roomFromRequest(w, r)
	if r2 == nil {
		return
	}
	s.RequireAuth(func(w http.ResponseWriter, r *http.Request, user *store.User) {
		var req struct {
			Show bool `json:"show"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		r2.Sg.Lock()
		g := r2.Sg.GetGame()
		err := g.ExposeCards(user.UUID, req.Show)
		if err != nil {
			r2.Sg.Unlock()
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		r2.Sg.Unlock()

		s.broadcast(r2)
		writeJSON(w, http.StatusOK, map[string]string{"message": "card exposure updated successfully"})
	})(w, r)
}
