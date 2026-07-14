package api

import (
	"net/http"

	"lil-poker/internal/middleware"
	"lil-poker/internal/store"
)

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	const bodyLimit = 32 * 1024

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("POST /api/auth/register", s.limit(s.limitBody(s.handleRegister, bodyLimit), s.authLimiter))
	mux.HandleFunc("POST /api/auth/guest", s.limit(s.limitBody(s.handleGuest, bodyLimit), s.authLimiter))
	mux.HandleFunc("POST /api/auth/guest/resume", s.limit(s.limitBody(s.handleGuestResume, bodyLimit), s.authLimiter))
	mux.HandleFunc("POST /api/auth/login", s.limit(s.limitBody(s.handleLogin, bodyLimit), s.authLimiter))
	mux.HandleFunc("GET /api/auth/me", s.handleMe)
	mux.HandleFunc("POST /api/auth/logout", s.handleLogout)
	mux.HandleFunc("POST /api/auth/rebuy", s.limit(s.limitBody(s.handleRebuy, bodyLimit), s.authLimiter))
	mux.HandleFunc("GET /api/rooms", s.handleListRooms)
	mux.HandleFunc("POST /api/rooms", s.limit(s.limitBody(s.handleCreateRoom, bodyLimit), s.generalLimiter))
	mux.HandleFunc("POST /api/game/create", s.limit(s.limitBody(s.handleCreate, bodyLimit), s.generalLimiter))
	mux.HandleFunc("POST /api/game/players", s.limit(s.limitBody(s.handleAddPlayer, bodyLimit), s.generalLimiter))
	mux.HandleFunc("POST /api/game/start", s.limit(s.limitBody(s.handleStart, bodyLimit), s.generalLimiter))
	mux.HandleFunc("POST /api/game/act", s.limit(s.limitBody(s.handleAct, bodyLimit), s.actionLimiter))
	mux.HandleFunc("POST /api/game/sit", s.limit(s.limitBody(s.handleSit, bodyLimit), s.generalLimiter))
	mux.HandleFunc("POST /api/game/stand", s.limit(s.limitBody(s.handleStand, bodyLimit), s.generalLimiter))
	mux.HandleFunc("POST /api/game/add-bot", s.limit(s.limitBody(s.handleAddBot, bodyLimit), s.generalLimiter))
	mux.HandleFunc("POST /api/game/blinds", s.limit(s.limitBody(s.handleUpdateBlinds, bodyLimit), s.generalLimiter))
	mux.HandleFunc("GET /api/game/status", s.handleStatus)
	mux.HandleFunc("GET /api/game/ws", s.handleWS)
	mux.HandleFunc("POST /api/game/show-cards", s.limit(s.limitBody(s.handleShowCards, bodyLimit), s.generalLimiter))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			if s.isOriginAllowed(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			origin := r.Header.Get("Origin")
			if origin != "" && !s.isOriginAllowed(origin) {
				writeError(w, http.StatusForbidden, "CSRF check failed: origin not allowed")
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}

func (s *Server) isOriginAllowed(origin string) bool {
	if len(s.allowedOrigins) == 0 {
		return true
	}
	for _, o := range s.allowedOrigins {
		if o == origin {
			return true
		}
	}
	return false
}

func (s *Server) RequireAuth(next func(w http.ResponseWriter, r *http.Request, user *store.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.bypassAuth {
			next(w, r, nil)
			return
		}
		user, err := s.getAuthenticatedUser(r)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "database error")
			return
		}
		if user == nil {
			writeError(w, http.StatusUnauthorized, "must be logged in")
			return
		}
		next(w, r, user)
	}
}

func (s *Server) limit(next http.HandlerFunc, limiter *middleware.Limiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.bypassAuth {
			next(w, r)
			return
		}
		identifier := r.RemoteAddr
		if cookie, err := r.Cookie("poker_session"); err == nil && cookie.Value != "" {
			identifier = cookie.Value
		}
		if !limiter.Allow(identifier) {
			writeError(w, http.StatusTooManyRequests, "too many requests, please try again later")
			return
		}
		next(w, r)
	}
}

func (s *Server) limitBody(next http.HandlerFunc, maxBytes int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		next(w, r)
	}
}
