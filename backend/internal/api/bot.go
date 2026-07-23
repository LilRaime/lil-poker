package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"lil-poker/internal/store"
)

func (s *Server) handleAddBot(w http.ResponseWriter, r *http.Request) {
	r2 := s.roomFromRequest(w, r)
	if r2 == nil {
		return
	}

	s.RequireAuth(func(w http.ResponseWriter, r *http.Request, user *store.User) {
		if user != nil && r2.CreatorID != "" && user.UUID != r2.CreatorID {
			writeError(w, http.StatusForbidden, "only the room creator can add bots")
			return
		}

		r2.Sg.Lock()
		playersCount := len(r2.Sg.GetGame().Players)
		maxPlayers := r2.MaxPlayers
		r2.Sg.Unlock()

		if playersCount >= maxPlayers {
			writeError(w, http.StatusBadRequest, "room is full")
			return
		}

		botName := fmt.Sprintf("RL_Bot_%d", time.Now().Unix()%1000)
		dockerSocketPath := "/var/run/docker.sock"

		if _, err := os.Stat(dockerSocketPath); err == nil {
			// Docker Mode
			hostname, hostErr := os.Hostname()
			if hostErr != nil {
				slog.Error("Failed to get hostname for Docker API lookup", "err", hostErr)
				writeError(w, http.StatusInternalServerError, "failed to get backend hostname")
				return
			}

			dialer := func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", dockerSocketPath)
			}
			client := &http.Client{
				Transport: &http.Transport{
					DialContext: dialer,
				},
			}

			resp, getErr := client.Get(fmt.Sprintf("http://localhost/containers/%s/json", hostname))
			if getErr != nil {
				slog.Error("Docker API lookup failed", "err", getErr)
				writeError(w, http.StatusInternalServerError, "failed to contact Docker daemon")
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				slog.Error("Docker API inspect returned non-200 status", "status", resp.Status)
				writeError(w, http.StatusInternalServerError, "failed to inspect self container")
				return
			}

			var inspectData struct {
				NetworkSettings struct {
					Networks map[string]interface{} `json:"Networks"`
				} `json:"NetworkSettings"`
			}
			if decodeErr := json.NewDecoder(resp.Body).Decode(&inspectData); decodeErr != nil {
				slog.Error("Failed to decode Docker API response", "err", decodeErr)
				writeError(w, http.StatusInternalServerError, "failed to parse container inspection")
				return
			}

			networkName := "lil-poker_default"
			for netName := range inspectData.NetworkSettings.Networks {
				networkName = netName
				break
			}

			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}
			backendURL := fmt.Sprintf("http://%s:%s", hostname, port)

			type HostConfigType struct {
				NetworkMode string `json:"NetworkMode"`
				AutoRemove  bool   `json:"AutoRemove"`
			}
			type CreateContainerReq struct {
				Image      string         `json:"Image"`
				Cmd        []string       `json:"Cmd"`
				HostConfig HostConfigType `json:"HostConfig"`
			}

			createPayload := CreateContainerReq{
				Image: "lil-poker-rl",
				Cmd: []string{
					"python", "-m", "agent.play_live",
					"--room", r2.ID,
					"--url", backendURL,
					"--name", botName,
					"--device", "cpu",
				},
				HostConfig: HostConfigType{
					NetworkMode: networkName,
					AutoRemove:  true,
				},
			}

			payloadBytes, marshalErr := json.Marshal(createPayload)
			if marshalErr != nil {
				slog.Error("Failed to marshal docker create request", "err", marshalErr)
				writeError(w, http.StatusInternalServerError, "failed to serialize container payload")
				return
			}

			createURL := fmt.Sprintf("http://localhost/containers/create?name=%s", url.QueryEscape(botName))
			postResp, createErr := client.Post(createURL, "application/json", bytes.NewReader(payloadBytes))
			if createErr != nil {
				slog.Error("Failed to post container create to Docker socket", "err", createErr)
				writeError(w, http.StatusInternalServerError, "failed to request container creation")
				return
			}
			defer postResp.Body.Close()

			if postResp.StatusCode == http.StatusNotFound {
				slog.Error("Docker image 'lil-poker-rl' not found")
				writeError(w, http.StatusBadRequest, "Docker image 'lil-poker-rl' not found. Please build it first: `docker build -t lil-poker-rl .`")
				return
			}

			if postResp.StatusCode != http.StatusCreated && postResp.StatusCode != http.StatusOK {
				slog.Error("Docker container create returned status", "status", postResp.Status)
				writeError(w, http.StatusInternalServerError, "Docker failed to create container")
				return
			}

			var createResp struct {
				ID string `json:"Id"`
			}
			if decodeErr := json.NewDecoder(postResp.Body).Decode(&createResp); decodeErr != nil {
				slog.Error("Failed to decode Docker container create response", "err", decodeErr)
				writeError(w, http.StatusInternalServerError, "failed to parse container creation response")
				return
			}

			startUrl := fmt.Sprintf("http://localhost/containers/%s/start", createResp.ID)
			startResp, startErr := client.Post(startUrl, "application/json", nil)
			if startErr != nil {
				slog.Error("Failed to start container via Docker API", "err", startErr)
				writeError(w, http.StatusInternalServerError, "failed to start bot container")
				return
			}
			defer startResp.Body.Close()

			if startResp.StatusCode != http.StatusNoContent && startResp.StatusCode != http.StatusOK {
				slog.Error("Docker container start returned status", "status", startResp.Status)
				writeError(w, http.StatusInternalServerError, "Docker failed to start container")
				return
			}

			slog.Info("Successfully spawned bot inside Docker container", "name", botName, "id", createResp.ID)
			writeJSON(w, http.StatusCreated, map[string]string{
				"message": fmt.Sprintf("Spawned bot '%s' inside Docker container", botName),
			})
			return
		}

		// Local fallback
		pythonPath := "/mnt/e/lil-poker-rl/.venv/bin/python"
		botDir := "/mnt/e/lil-poker-rl"

		if _, err := os.Stat(pythonPath); os.IsNotExist(err) {
			cwd, _ := os.Getwd()
			relPath := filepath.Join(cwd, "..", "lil-poker-rl")
			if _, err := os.Stat(filepath.Join(relPath, ".venv", "bin", "python")); err == nil {
				pythonPath = filepath.Join(relPath, ".venv", "bin", "python")
				botDir = relPath
			} else if _, err := os.Stat(filepath.Join(cwd, "..", "..", "lil-poker-rl", ".venv", "bin", "python")); err == nil {
				pythonPath = filepath.Join(cwd, "..", "..", "lil-poker-rl", ".venv", "bin", "python")
				botDir = filepath.Join(cwd, "..", "..", "lil-poker-rl")
			} else {
				pythonPath = "python3"
				botDir = "."
			}
		}

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		backendURL := fmt.Sprintf("http://localhost:%s", port)

		cmdArgs := []string{
			"-m", "agent.play_live",
			"--room", r2.ID,
			"--url", backendURL,
			"--name", botName,
			"--device", "cpu",
		}

		cmd := exec.Command(pythonPath, cmdArgs...)
		cmd.Dir = botDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		startErr := cmd.Start()
		if startErr != nil {
			slog.Error("Failed to start local bot process", "err", startErr)
			writeError(w, http.StatusInternalServerError, "failed to start bot locally")
			return
		}

		go func() {
			waitErr := cmd.Wait()
			if waitErr != nil {
				slog.Info("Local bot process exited", "name", botName, "err", waitErr)
			} else {
				slog.Info("Local bot process exited cleanly", "name", botName)
			}
		}()

		slog.Info("Successfully spawned local bot process", "name", botName)
		writeJSON(w, http.StatusCreated, map[string]string{
			"message": fmt.Sprintf("Spawned local bot process '%s'", botName),
		})
	})(w, r)
}
