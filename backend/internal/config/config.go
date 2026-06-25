package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	SmallBlind     int
	BigBlind       int
	CookieSecret   string
	TurnTimeout    time.Duration
	AllowedOrigins []string
	RedisAddr      string
}

func LoadConfig() Config {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "lil-poker"
	}

	sbStr := os.Getenv("SMALL_BLIND")
	bbStr := os.Getenv("BIG_BLIND")
	sb := 10
	bb := 20
	if val, err := strconv.Atoi(sbStr); err == nil && val > 0 {
		sb = val
	}
	if val, err := strconv.Atoi(bbStr); err == nil && val > 0 {
		bb = val
	}

	turnTimeout := 20 * time.Second
	if val, err := strconv.Atoi(os.Getenv("TURN_TIMEOUT_SECS")); err == nil && val > 0 {
		turnTimeout = time.Duration(val) * time.Second
	}

	var allowedOrigins []string
	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		for _, o := range strings.Split(raw, ",") {
			if o = strings.TrimSpace(o); o != "" {
				allowedOrigins = append(allowedOrigins, o)
			}
		}
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	return Config{
		DBHost:         host,
		DBPort:         port,
		DBUser:         user,
		DBPassword:     password,
		DBName:         dbname,
		SmallBlind:     sb,
		BigBlind:       bb,
		CookieSecret:   os.Getenv("COOKIE_SECRET"),
		TurnTimeout:    turnTimeout,
		AllowedOrigins: allowedOrigins,
		RedisAddr:      redisAddr,
	}
}
