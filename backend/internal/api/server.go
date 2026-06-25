package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"lil-poker/internal/config"
	"lil-poker/internal/middleware"
	"lil-poker/internal/room"
	"lil-poker/internal/store"
)

type Server struct {
	rm             *room.RoomManager
	upgrader       websocket.Upgrader
	db             *sql.DB
	rdb            *redis.Client
	authLimiter    *middleware.Limiter
	actionLimiter  *middleware.Limiter
	generalLimiter *middleware.Limiter
	cookieSecret   []byte
	janitorDone    chan struct{}
	bypassAuth     bool
	turnTimeout    time.Duration
	allowedOrigins []string
}

func NewServer(db *sql.DB, cfg config.Config) *Server {
	var secretBytes []byte
	if cfg.CookieSecret != "" {
		secretBytes = []byte(cfg.CookieSecret)
	} else {
		secretBytes = make([]byte, 32)
		_, _ = rand.Read(secretBytes)
	}

	turnTimeout := cfg.TurnTimeout
	if turnTimeout <= 0 {
		turnTimeout = 20 * time.Second
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	s := &Server{
		rm: room.NewRoomManager(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				if len(cfg.AllowedOrigins) == 0 {
					return true
				}
				origin := r.Header.Get("Origin")
				for _, o := range cfg.AllowedOrigins {
					if o == origin {
						return true
					}
				}
				return false
			},
		},
		db:             db,
		rdb:            rdb,
		authLimiter:    middleware.NewLimiter(0.5, 5),
		actionLimiter:  middleware.NewLimiter(5.0, 10),
		generalLimiter: middleware.NewLimiter(2.0, 5),
		cookieSecret:   secretBytes,
		turnTimeout:    turnTimeout,
		allowedOrigins: cfg.AllowedOrigins,
	}

	s.startCacheJanitor(1 * time.Minute)
	return s
}

func (s *Server) SetBypassAuth(v bool) {
	s.bypassAuth = v
}

func (s *Server) Close() {
	close(s.janitorDone)
	s.rm.Close()
	s.authLimiter.Close()
	s.actionLimiter.Close()
	s.generalLimiter.Close()
	s.rdb.Close()
}

func (s *Server) startCacheJanitor(interval time.Duration) {
	s.janitorDone = make(chan struct{})
	guestTicker := time.NewTicker(1 * time.Hour)
	go func() {
		defer guestTicker.Stop()
		for {
			select {
			case <-guestTicker.C:
				n, err := store.DeleteStaleGuests(s.db, 24*time.Hour)
				if err != nil {
					slog.Error("guest janitor failed", "err", err)
				} else if n > 0 {
					slog.Info("guest janitor removed stale guests", "count", n)
				}
			case <-s.janitorDone:
				return
			}
		}
	}()
}

func (s *Server) RoomManager() *room.RoomManager {
	return s.rm
}

func (s *Server) cacheSet(uuid string, user *store.User) {
	data, err := json.Marshal(user)
	if err != nil {
		slog.Error("failed to marshal user for cache", "err", err)
		return
	}
	ctx := context.Background()
	err = s.rdb.Set(ctx, "session:"+uuid, data, 5*time.Minute).Err()
	if err != nil {
		slog.Error("failed to write user session to redis", "uuid", uuid, "err", err)
	}
}

func (s *Server) cacheGet(uuid string) (*store.User, bool) {
	ctx := context.Background()
	data, err := s.rdb.Get(ctx, "session:"+uuid).Bytes()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		slog.Error("failed to get user session from redis", "uuid", uuid, "err", err)
		return nil, false
	}
	var u store.User
	if err := json.Unmarshal(data, &u); err != nil {
		slog.Error("failed to unmarshal user from redis session", "uuid", uuid, "err", err)
		return nil, false
	}
	return &u, true
}

func (s *Server) cacheInvalidate(uuid string) {
	ctx := context.Background()
	err := s.rdb.Del(ctx, "session:"+uuid).Err()
	if err != nil {
		slog.Error("failed to invalidate user session in redis", "uuid", uuid, "err", err)
	}
}

func (s *Server) updateUserChips(uuid string, chips int) error {
	err := store.UpdateUserChips(s.db, uuid, chips)
	s.cacheInvalidate(uuid)
	return err
}

func (s *Server) updateUserChipsAndRebuys(uuid string, chips, rebuys int) error {
	err := store.UpdateUserChipsAndRebuys(s.db, uuid, chips, rebuys)
	s.cacheInvalidate(uuid)
	return err
}

func (s *Server) updateUsersChipsTx(updates map[string]int) error {
	err := store.UpdateUsersChipsTx(s.db, updates)
	for uuid := range updates {
		s.cacheInvalidate(uuid)
	}
	return err
}

func (s *Server) flushChipsAsync(updates map[string]int) {
	if len(updates) == 0 {
		return
	}
	go func() {
		const maxAttempts = 3
		delay := time.Second
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			err := s.updateUsersChipsTx(updates)
			if err == nil {
				return
			}
			slog.Error("flushChipsAsync failed", "attempt", attempt, "max", maxAttempts, "err", err)
			if attempt < maxAttempts {
				time.Sleep(delay)
				delay *= 2
			}
		}
		slog.Error("flushChipsAsync: all attempts failed — chip data may be inconsistent", "max_attempts", maxAttempts)
	}()
}
