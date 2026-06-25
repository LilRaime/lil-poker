package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"lil-poker/internal/game"
	"lil-poker/internal/store"
)

const bcryptCost = 12

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func randomPassword() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Server) setSessionCookie(w http.ResponseWriter, r *http.Request, uuid string) {
	secure := r.TLS != nil || strings.ToLower(r.Header.Get("X-Forwarded-Proto")) == "https"
	http.SetCookie(w, &http.Cookie{
		Name:     "poker_session",
		Value:    s.signCookieValue(uuid),
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})
}

func (s *Server) getAuthenticatedUser(r *http.Request) (*store.User, error) {
	cookie, err := r.Cookie("poker_session")
	if err != nil {
		return nil, nil
	}
	uuid, ok := s.verifyCookieValue(cookie.Value)
	if !ok || uuid == "" {
		return nil, nil
	}

	if user, found := s.cacheGet(uuid); found {
		return user, nil
	}

	user, err := store.GetUserByUUID(s.db, uuid)
	if err != nil {
		slog.Error("GetUserByUUID failed", "uuid", uuid, "err", err)
		return nil, err
	}
	if user != nil {
		s.cacheSet(uuid, user)
	}
	return user, nil
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	user, err := s.getAuthenticatedUser(r)
	if err != nil {
		slog.Error("getAuthenticatedUser failed", "handler", "handleMe", "err", err)
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"uuid":     user.UUID,
		"username": user.Username,
		"chips":    user.Chips,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	secure := r.TLS != nil || strings.ToLower(r.Header.Get("X-Forwarded-Proto")) == "https"
	http.SetCookie(w, &http.Cookie{
		Name:     "poker_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || len(req.Username) < 3 {
		writeError(w, http.StatusBadRequest, "username must be at least 3 characters")
		return
	}
	if len(req.Username) > 20 {
		writeError(w, http.StatusBadRequest, "username must be at most 20 characters")
		return
	}
	if len(req.Password) < 4 {
		writeError(w, http.StatusBadRequest, "password must be at least 4 characters")
		return
	}

	existing, err := store.GetUserByUsername(s.db, req.Username)
	if err != nil {
		slog.Error("GetUserByUsername failed", "handler", "handleRegister", "username", req.Username, "err", err)
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if existing != nil {
		writeError(w, http.StatusBadRequest, "username is already taken")
		return
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		slog.Error("hashPassword failed", "handler", "handleRegister", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	user, err := store.CreateUser(s.db, req.Username, passwordHash)
	if err != nil {
		slog.Error("CreateUser failed", "handler", "handleRegister", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	s.setSessionCookie(w, r, user.UUID)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"uuid":     user.UUID,
		"username": user.Username,
		"chips":    user.Chips,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := store.GetUserByUsername(s.db, req.Username)
	if err != nil {
		slog.Error("GetUserByUsername failed", "handler", "handleLogin", "username", req.Username, "err", err)
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	if !checkPassword(req.Password, user.Password) {
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	s.setSessionCookie(w, r, user.UUID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"uuid":     user.UUID,
		"username": user.Username,
		"chips":    user.Chips,
	})
}

func (s *Server) handleRebuy(w http.ResponseWriter, r *http.Request) {
	s.RequireAuth(func(w http.ResponseWriter, r *http.Request, user *store.User) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		if user == nil {
			writeError(w, http.StatusUnauthorized, "must be logged in")
			return
		}

		roomID := r.URL.Query().Get("room")
		if roomID != "" {
			r2, ok := s.rm.GetRoom(roomID)
			if !ok {
				writeError(w, http.StatusNotFound, "room not found")
				return
			}

			r2.Sg.Lock()
			g := r2.Sg.GetGame()
			var player *game.Player
			for _, p := range g.Players {
				if p.ID == user.UUID {
					player = p
					break
				}
			}

			if player == nil {
				r2.Sg.Unlock()
				writeError(w, http.StatusBadRequest, "player not seated at this table")
				return
			}

			if player.Chips > 0 {
				r2.Sg.Unlock()
				writeError(w, http.StatusBadRequest, "rebuy is only allowed when you have 0 chips")
				return
			}

			if player.RebuysRemaining <= 0 {
				r2.Sg.Unlock()
				writeError(w, http.StatusBadRequest, "no rebuys remaining")
				return
			}

			player.RebuysRemaining--
			rebuyAmount := r2.StartingChips
			if rebuyAmount <= 0 {
				rebuyAmount = 1000
			}

			player.Chips = rebuyAmount
			player.SittingOut = false
			r2.Sg.Unlock()

			if r2.StartingChips == 0 {
				_ = s.updateUserChipsAndRebuys(user.UUID, rebuyAmount, player.RebuysRemaining)
			}

			s.broadcast(r2)

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"uuid":             user.UUID,
				"username":         user.Username,
				"chips":            rebuyAmount,
				"rebuys_remaining": player.RebuysRemaining,
			})
			return
		}

		if user.Chips > 0 {
			writeError(w, http.StatusBadRequest, "rebuy is only allowed when you have 0 chips")
			return
		}

		rebuys := user.RebuysRemaining
		if rebuys <= 0 {
			rebuys = 3
		}

		err := s.updateUserChipsAndRebuys(user.UUID, 1000, rebuys-1)
		if err != nil {
			slog.Error("updateUserChipsAndRebuys failed", "uuid", user.UUID, "err", err)
			writeError(w, http.StatusInternalServerError, "failed to update chips")
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"uuid":             user.UUID,
			"username":         user.Username,
			"chips":            1000,
			"rebuys_remaining": rebuys - 1,
		})
	})(w, r)
}

func (s *Server) handleGuest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || len(req.Username) < 3 {
		writeError(w, http.StatusBadRequest, "nickname must be at least 3 characters")
		return
	}
	if len(req.Username) > 12 {
		writeError(w, http.StatusBadRequest, "nickname must be at most 12 characters")
		return
	}

	baseName := req.Username
	finalName := baseName
	importTime := time.Now().UnixNano()
	const maxGuestNameAttempts = 100
	for counter := 0; counter < maxGuestNameAttempts; counter++ {
		existing, err := store.GetUserByUsername(s.db, finalName)
		if err != nil {
			slog.Error("GetUserByUsername failed", "handler", "handleGuest", "err", err)
			writeError(w, http.StatusInternalServerError, "database error")
			return
		}
		if existing == nil {
			break
		}
		if counter == maxGuestNameAttempts-1 {
			slog.Error("handleGuest: could not find unique name", "base", baseName)
			writeError(w, http.StatusInternalServerError, "failed to create guest user")
			return
		}
		suffix := (importTime + int64(counter)) % 10000
		if suffix < 0 {
			suffix = -suffix
		}
		finalName = fmt.Sprintf("%s_%d", baseName, suffix)
	}

	rawPass, err := randomPassword()
	if err != nil {
		slog.Error("randomPassword failed", "handler", "handleGuest", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create guest user")
		return
	}
	passwordHash, err := hashPassword(rawPass)
	if err != nil {
		slog.Error("hashPassword failed", "handler", "handleGuest", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create guest user")
		return
	}
	user, err := store.CreateGuestUser(s.db, finalName, passwordHash)
	if err != nil {
		slog.Error("CreateUser failed for guest", "handler", "handleGuest", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create guest user")
		return
	}

	s.setSessionCookie(w, r, user.UUID)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"uuid":     user.UUID,
		"username": user.Username,
		"chips":    user.Chips,
		"is_guest": true,
		"password": rawPass,
	})
}

func (s *Server) handleGuestResume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		UUID     string `json:"uuid"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UUID == "" {
		writeError(w, http.StatusBadRequest, "uuid is required")
		return
	}

	user, err := store.GetUserByUUID(s.db, req.UUID)
	if err != nil {
		slog.Error("GetUserByUUID failed", "handler", "handleGuestResume", "err", err)
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if user == nil || !user.IsGuest {
		writeError(w, http.StatusNotFound, "guest session not found")
		return
	}

	if !checkPassword(req.Password, user.Password) {
		writeError(w, http.StatusUnauthorized, "invalid guest credentials")
		return
	}

	s.setSessionCookie(w, r, user.UUID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"uuid":     user.UUID,
		"username": user.Username,
		"chips":    user.Chips,
		"is_guest": true,
	})
}
