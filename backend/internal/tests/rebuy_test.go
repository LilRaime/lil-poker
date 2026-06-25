package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"lil-poker/internal/api"
	"lil-poker/internal/store"
)

func TestRebuy(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	server := NewServer(db, LoadConfig())
	handler := server.Handler()

	roomID := createTestRoom(t, handler)
	uuidStr := registerAndSeatPlayer(t, handler, "gungnir", roomID)

	_ = UpdateUserChips(db, uuidStr, 0)

	room, _ := server.RoomManager().GetRoom(roomID)
	room.Sg.Lock()
	for _, p := range room.Sg.GetGame().Players {
		if p.ID == uuidStr {
			p.Chips = 0
		}
	}
	room.Sg.Unlock()

	rebReq := RebuyRequest{UUID: uuidStr}
	body, _ := json.Marshal(rebReq)
	req := httptest.NewRequest("POST", "/api/auth/rebuy", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var res map[string]interface{}
	json.NewDecoder(w.Body).Decode(&res)
	if res["chips"].(float64) != float64(rebuyAmount) {
		t.Errorf("expected %d chips, got %v", rebuyAmount, res["chips"])
	}

	user, _ := GetUserByUUID(db, uuidStr)
	if user.Chips != rebuyAmount {
		t.Errorf("expected DB chips to be %d, got %d", rebuyAmount, user.Chips)
	}

	room.Sg.Lock()
	for _, p := range room.Sg.GetGame().Players {
		if p.ID == uuidStr {
			if p.Chips != rebuyAmount {
				t.Errorf("expected player chips at table to be %d, got %d", rebuyAmount, p.Chips)
			}
		}
	}
	room.Sg.Unlock()
}

func TestRebuy_DeadlockReset(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	server := NewServer(db, LoadConfig())
	handler := server.Handler()

	regReq := api.AuthRequest{Username: "deadlocked", Password: "testpassword"}
	bodyReg, _ := json.Marshal(regReq)
	reqReg := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(bodyReg))
	wReg := httptest.NewRecorder()
	handler.ServeHTTP(wReg, reqReg)

	var regRes map[string]interface{}
	json.NewDecoder(wReg.Body).Decode(&regRes)
	uuidStr := regRes["uuid"].(string)

	_ = store.UpdateUserChipsAndRebuys(db, uuidStr, 0, 0)

	rebReq := RebuyRequest{UUID: uuidStr}
	body, _ := json.Marshal(rebReq)
	req := httptest.NewRequest("POST", "/api/auth/rebuy", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for reset rebuy, got %d. Body: %s", w.Code, w.Body.String())
	}

	var res map[string]interface{}
	json.NewDecoder(w.Body).Decode(&res)
	if res["chips"].(float64) != float64(rebuyAmount) {
		t.Errorf("expected %d chips, got %v", rebuyAmount, res["chips"])
	}
	if res["rebuys_remaining"].(float64) != 2 {
		t.Errorf("expected 2 rebuys remaining, got %v", res["rebuys_remaining"])
	}

	user, _ := store.GetUserByUUID(db, uuidStr)
	if user.Chips != rebuyAmount || user.RebuysRemaining != 2 {
		t.Errorf("expected DB to have %d chips and 2 rebuys, got chips=%d, rebuys=%d", rebuyAmount, user.Chips, user.RebuysRemaining)
	}
}
