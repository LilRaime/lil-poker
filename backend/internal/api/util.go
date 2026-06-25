package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"lil-poker/internal/game"
	"lil-poker/internal/types"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, errMsg string) {
	writeJSON(w, status, map[string]string{"error": errMsg})
}

func parseAction(actStr string) (game.Action, error) {
	switch strings.ToLower(actStr) {
	case "fold":
		return game.ActionFold, nil
	case "check":
		return game.ActionCheck, nil
	case "call":
		return game.ActionCall, nil
	case "raise":
		return game.ActionRaise, nil
	case "all_in", "allin":
		return game.ActionAllIn, nil
	default:
		return 0, fmt.Errorf("invalid action: %s", actStr)
	}
}

func (s *Server) signCookieValue(val string) string {
	mac := hmac.New(sha256.New, s.cookieSecret)
	mac.Write([]byte(val))
	signature := hex.EncodeToString(mac.Sum(nil))
	return val + "." + signature
}

func (s *Server) verifyCookieValue(signedVal string) (string, bool) {
	parts := strings.SplitN(signedVal, ".", 2)
	if len(parts) != 2 {
		return "", false
	}
	val, signature := parts[0], parts[1]

	mac := hmac.New(sha256.New, s.cookieSecret)
	mac.Write([]byte(val))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) == 1 {
		return val, true
	}
	return "", false
}

type GameStateResponse = types.GameStateResponse

var _ types.WSMessage
