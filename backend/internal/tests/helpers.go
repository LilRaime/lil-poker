package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"lil-poker/internal/api"
	"lil-poker/internal/config"
	"lil-poker/internal/room"
	"lil-poker/internal/store"
)

type User = store.User
type RoomInfo = room.RoomInfo

type Config = config.Config
type Server = api.Server

const rebuyAmount = 1000

type CreateRoomRequest = api.CreateRoomRequest
type CreateGameRequest = api.CreateGameRequest
type AddPlayerRequest = api.AddPlayerRequest
type ActRequest = api.ActRequest
type AuthRequest = api.AuthRequest
type RebuyRequest = api.RebuyRequest
type SitRequest = api.SitRequest
type UpdateBlindsRequest = api.UpdateBlindsRequest
type GameStateResponse = api.GameStateResponse

func NewServer(db *sql.DB, cfg config.Config) *Server {
	srv := api.NewServer(db, cfg)
	srv.SetBypassAuth(true)
	return srv
}

func LoadConfig() config.Config {
	return config.LoadConfig()
}

func UpdateUserChips(db *sql.DB, uuid string, chips int) error {
	return store.UpdateUserChips(db, uuid, chips)
}

func GetUserByUUID(db *sql.DB, uuid string) (*store.User, error) {
	return store.GetUserByUUID(db, uuid)
}

func setupTestDB(t *testing.T) *sql.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "postgres"
	}
	if password == "" {
		password = "postgres"
	}

	connStrDefault := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable connect_timeout=2",
		host, port, user, password)

	dbDefault, err := sql.Open("postgres", connStrDefault)
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL is not running. Error: %v", err)
		return nil
	}
	if err := dbDefault.Ping(); err != nil {
		dbDefault.Close()
		t.Skipf("Skipping test: PostgreSQL is not running. Error: %v", err)
		return nil
	}

	_, _ = dbDefault.Exec("CREATE DATABASE poker_test")
	dbDefault.Close()

	connStrTest := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=poker_test sslmode=disable connect_timeout=2",
		host, port, user, password)
	dbTest, err := sql.Open("postgres", connStrTest)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := dbTest.Ping(); err != nil {
		dbTest.Close()
		t.Fatalf("Failed to ping test database: %v", err)
	}

	err = store.RunMigrations(dbTest)
	if err != nil {
		dbTest.Close()
		t.Fatalf("Failed to init schema on test db: %v", err)
	}

	_, _ = dbTest.Exec("DELETE FROM users")
	return dbTest
}

func createTestRoom(t *testing.T, handler http.Handler) string {
	body, _ := json.Marshal(api.CreateRoomRequest{Name: "Test Room", SmallBlind: 5, BigBlind: 10})
	req := httptest.NewRequest("POST", "/api/rooms", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create room: status %d, body: %s", w.Code, w.Body.String())
	}
	var ri room.RoomInfo
	if err := json.NewDecoder(w.Body).Decode(&ri); err != nil {
		t.Fatalf("failed to decode room info: %v", err)
	}
	return ri.ID
}

func registerAndSeatPlayer(t *testing.T, handler http.Handler, username string, roomID string) string {
	regReq := api.AuthRequest{Username: username, Password: "testpassword"}
	bodyReg, _ := json.Marshal(regReq)
	reqReg := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(bodyReg))
	wReg := httptest.NewRecorder()
	handler.ServeHTTP(wReg, reqReg)

	if wReg.Code != http.StatusCreated {
		t.Fatalf("failed to register player %s: status %d, body: %s", username, wReg.Code, wReg.Body.String())
	}

	var regRes map[string]interface{}
	if err := json.NewDecoder(wReg.Body).Decode(&regRes); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}
	uuidStr := regRes["uuid"].(string)

	seatReq := api.AddPlayerRequest{UUID: uuidStr}
	bodySeat, _ := json.Marshal(seatReq)
	reqSeat := httptest.NewRequest("POST", "/api/game/players?room="+roomID, bytes.NewBuffer(bodySeat))
	wSeat := httptest.NewRecorder()
	handler.ServeHTTP(wSeat, reqSeat)

	if wSeat.Code != http.StatusCreated {
		t.Fatalf("failed to seat player %s: status %d, body: %s", username, wSeat.Code, wSeat.Body.String())
	}

	return uuidStr
}
