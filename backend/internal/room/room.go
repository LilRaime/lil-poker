package room

import (
	"crypto/rand"
	"math/big"
	"sync"
	"time"
)

const roomCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
const roomCodeLen = 6
const RoomTTL = 2 * time.Hour

type RoomInfo struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	PlayerCount         int    `json:"player_count"`
	MaxPlayers          int    `json:"max_players"`
	Phase               string `json:"phase"`
	SmallBlind          int    `json:"small_blind"`
	BigBlind            int    `json:"big_blind"`
	CreatorID           string `json:"creator_id"`
	BlindEscalationMins int    `json:"blind_escalation_mins"`
	StartingChips       int    `json:"starting_chips"`
	MaxRebuys           int    `json:"max_rebuys"`
	TurnTimeoutSecs     int    `json:"turn_timeout_secs"`
}

type Room struct {
	ID                  string
	Name                string
	CreatorID           string
	MaxPlayers          int
	Sg                  *SafeGame
	Wsm                 *WSManager
	TimerCancel         chan struct{}
	TimerMu             sync.Mutex
	LastActive          time.Time
	Mu                  sync.Mutex
	BlindEscalationMins int
	StartingChips       int
	MaxRebuys           int
	TurnTimeoutSecs     int
}

func newRoom(id, name, creatorID string, maxPlayers, sb, bb, escalationMins, startingChips, maxRebuys, turnTimeoutSecs int) *Room {
	sg := NewSafeGame(sb, bb)
	sg.CreatorID = creatorID
	if maxPlayers < 2 || maxPlayers > 8 {
		maxPlayers = 8
	}
	r := &Room{
		ID:                  id,
		Name:                name,
		CreatorID:           creatorID,
		MaxPlayers:          maxPlayers,
		Sg:                  sg,
		Wsm:                 NewWSManager(sg, startingChips),
		LastActive:          time.Now(),
		BlindEscalationMins: escalationMins,
		StartingChips:       startingChips,
		MaxRebuys:           maxRebuys,
		TurnTimeoutSecs:     turnTimeoutSecs,
	}
	sg.game.BlindEscalationMins = escalationMins
	sg.game.MaxRebuys = maxRebuys
	return r
}

func (room *Room) Touch() {
	room.Mu.Lock()
	room.LastActive = time.Now()
	room.Mu.Unlock()
}

func (room *Room) Close() {
	room.TimerMu.Lock()
	if room.TimerCancel != nil {
		close(room.TimerCancel)
		room.TimerCancel = nil
	}
	room.TimerMu.Unlock()
	room.Wsm.Close()
}

func (room *Room) Info() RoomInfo {
	snap := room.Sg.getSnapshot()
	return RoomInfo{
		ID:                  room.ID,
		Name:                room.Name,
		PlayerCount:         len(snap.Players),
		MaxPlayers:          room.MaxPlayers,
		Phase:               snap.Phase,
		SmallBlind:          snap.SmallBlind,
		BigBlind:            snap.BigBlind,
		CreatorID:           room.CreatorID,
		BlindEscalationMins: room.BlindEscalationMins,
		StartingChips:       room.StartingChips,
		MaxRebuys:           room.MaxRebuys,
		TurnTimeoutSecs:     room.TurnTimeoutSecs,
	}
}

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
	done  chan struct{}
}

func NewRoomManager() *RoomManager {
	rm := &RoomManager{
		rooms: make(map[string]*Room),
		done:  make(chan struct{}),
	}
	go rm.janitor()
	return rm
}

func (rm *RoomManager) Close() {
	close(rm.done)
}

func (rm *RoomManager) generateID() string {
	for {
		b := make([]byte, roomCodeLen)
		for i := range b {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(roomCodeChars))))
			if err != nil {
				n = big.NewInt(0)
			}
			b[i] = roomCodeChars[n.Int64()]
		}
		id := string(b)
		if _, exists := rm.rooms[id]; !exists {
			return id
		}
	}
}

func (rm *RoomManager) CreateRoom(name, creatorID string, maxPlayers, sb, bb, escalationMins, startingChips, maxRebuys, turnTimeoutSecs int) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	id := rm.generateID()
	r := newRoom(id, name, creatorID, maxPlayers, sb, bb, escalationMins, startingChips, maxRebuys, turnTimeoutSecs)
	rm.rooms[id] = r
	return r
}

func (rm *RoomManager) GetRoom(id string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	r, ok := rm.rooms[id]
	return r, ok
}

func (rm *RoomManager) DeleteRoom(id string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if r, ok := rm.rooms[id]; ok {
		r.Close()
		delete(rm.rooms, id)
	}
}

func (rm *RoomManager) ListRooms() []RoomInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	list := make([]RoomInfo, 0, len(rm.rooms))
	for _, r := range rm.rooms {
		list = append(list, r.Info())
	}
	return list
}

func (rm *RoomManager) janitor() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-rm.done:
			return
		case <-ticker.C:
			now := time.Now()
			var toDelete []string
			var toClose []*Room

			rm.mu.RLock()
			for id, r := range rm.rooms {
				r.Mu.Lock()
				idle := now.Sub(r.LastActive)
				r.Mu.Unlock()

				snap := r.Sg.getSnapshot()
				if len(snap.Players) == 0 && idle > RoomTTL {
					toDelete = append(toDelete, id)
					toClose = append(toClose, r)
				}
			}
			rm.mu.RUnlock()

			if len(toDelete) > 0 {
				rm.mu.Lock()
				for _, id := range toDelete {
					delete(rm.rooms, id)
				}
				rm.mu.Unlock()

				for _, r := range toClose {
					go r.Close()
				}
			}
		}
	}
}
