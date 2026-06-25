package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocketBroadcast(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	server := NewServer(db, LoadConfig())
	testServer := httptest.NewServer(server.Handler())
	defer testServer.Close()

	bodyRoom, _ := json.Marshal(CreateRoomRequest{Name: "WS Test Room", SmallBlind: 5, BigBlind: 10})
	respRoom, err := http.Post(testServer.URL+"/api/rooms", "application/json", bytes.NewBuffer(bodyRoom))
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}
	defer respRoom.Body.Close()
	var roomInfo RoomInfo
	if err := json.NewDecoder(respRoom.Body).Decode(&roomInfo); err != nil {
		t.Fatalf("failed to decode room info: %v", err)
	}
	roomID := roomInfo.ID

	wsURL := "ws" + testServer.URL[4:] + "/api/game/ws?room=" + roomID

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect to websocket: %v", err)
	}
	defer conn.Close()

	var initialMsg GameStateResponse
	err = conn.ReadJSON(&initialMsg)
	if err != nil {
		t.Fatalf("failed to read initial message: %v", err)
	}
	if initialMsg.Phase != "Waiting" {
		t.Errorf("expected initial phase to be Waiting, got %s", initialMsg.Phase)
	}

	regReq := AuthRequest{Username: "gungnir", Password: "testpassword"}
	bodyReg, _ := json.Marshal(regReq)
	respReg, err := http.Post(testServer.URL+"/api/auth/register", "application/json", bytes.NewBuffer(bodyReg))
	if err != nil {
		t.Fatalf("failed to register player: %v", err)
	}
	defer respReg.Body.Close()
	if respReg.StatusCode != http.StatusCreated {
		t.Fatalf("expected StatusCreated, got %d", respReg.StatusCode)
	}
	var regRes map[string]interface{}
	if err := json.NewDecoder(respReg.Body).Decode(&regRes); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}
	uuidStr := regRes["uuid"].(string)

	seatReq := AddPlayerRequest{UUID: uuidStr}
	bodySeat, _ := json.Marshal(seatReq)
	respSeat, err := http.Post(testServer.URL+"/api/game/players?room="+roomID, "application/json", bytes.NewBuffer(bodySeat))
	if err != nil {
		t.Fatalf("failed to add player: %v", err)
	}
	defer respSeat.Body.Close()
	if respSeat.StatusCode != http.StatusCreated {
		t.Fatalf("expected StatusCreated, got %d", respSeat.StatusCode)
	}

	var updateMsg GameStateResponse
	err = conn.ReadJSON(&updateMsg)
	if err != nil {
		t.Fatalf("failed to read broadcast message: %v", err)
	}

	if len(updateMsg.Players) != 1 {
		t.Errorf("expected 1 player in update message, got %d", len(updateMsg.Players))
	}
	if updateMsg.Players[0].Name != "gungnir" {
		t.Errorf("expected player name to be 'gungnir', got %s", updateMsg.Players[0].Name)
	}
}
