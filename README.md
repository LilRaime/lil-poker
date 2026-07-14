# 🃏 Lil Poker

**Lil Poker** is a web-based multiplayer Texas Hold'em poker game played in real-time. Built with a modern and highly efficient technology stack, it allows players to create custom game rooms with flexible settings, chat, and play poker with friends.

## 📸 Screenshots

<p align="center">
  <img src="screenshots/1.png" alt="Lil Poker Login/Authentication" width="49%" />
  <img src="screenshots/2.png" alt="Lil Poker Room Browser / Lobby" width="49%" />
</p>
<p align="center">
  <img src="screenshots/3.png" alt="Lil Poker Game Room Table" width="49%" />
  <img src="screenshots/4.png" alt="Lil Poker Gameplay in Action" width="49%" />
</p>

---

## 🚀 Key Features

- **Real-Time Gameplay**: Instant game state updates and synchronization powered by **WebSockets**.
- **Room Management (Lobby)**:
  - Create custom game rooms with tailored rules.
  - Adjust small/big blind values and turn timeout durations.
  - Time-based automatic blind escalation.
  - Set initial stack sizes and maximum rebuy limits.
  - **Chip Mode** — choose between:
    - 🏆 **Tournament** — every player starts with a fixed chip stack (isolated per room).
    - 💰 **Persistent** — players bring their real account balance; winnings and losses carry over across sessions.
- **Flexible Authentication**:
  - Quick Guest login to start playing instantly without registration.
  - Registered user accounts to persist chip balances and records.
- **Interactive Game Interface**:
  - Responsive 3D-like table layout showing seats, chip stacks, dealer button, board cards, and sub-pots.
  - Action panel with Fold, Check/Call, and Bet/Raise options (featuring an interactive bet slider).
  - Turn countdown timer for active players.
  - Hand history tracking.
  - Optional auto-rebuy toggle.
  - Chip mode badge (Tournament / Persistent) visible in both the lobby and the game header.
- **Communication & Logging**: Integrated real-time chat and system logs highlighting active table events.
- **Security & Stability**:
  - Session collision detection (disallows multiple active tabs for the same account).
  - Rate limiting on authentication, room creation, and actions.
  - Strict Origin (CORS/CSRF) checks.

---

## 🛠 Technology Stack

### Backend
- **Programming Language**: Go (Golang) v1.26
- **Database**: PostgreSQL 17 (persisting user profiles and chip balances)
- **Cache**: Redis 8.8 (caching and validating session cookies)
- **Real-time Engine**: Gorilla WebSockets
- **Migrations**: Automated sql-migrations embedded directly in the binary (`go:embed`)

### Frontend
- **Framework/Library**: React (TypeScript)
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **Communication**: Native WebSocket API & Fetch API

### Infrastructure & DevOps
- **Containerization**: Docker & Docker Compose
- **Kubernetes**: Kind + Helm chart (see `.infra/helm/`)
- **Automation**: Makefile

---

## 📂 Project Structure

```text
lil-poker/
├── backend/               # Go backend server and game engine
│   ├── cmd/server/        # Application entrypoint (main.go)
│   └── internal/          # Core business logic packages
│       ├── api/           # HTTP endpoints, routing, WebSocket handlers
│       ├── card/          # Card representations (suits, ranks)
│       ├── deck/          # Deck operations and shuffling
│       ├── game/          # Texas Hold'em game engine state machine
│       ├── hand/          # Hand strength evaluator
│       ├── middleware/    # Rate limiters and utility middlewares
│       ├── room/          # Rooms manager and client broadcaster
│       ├── store/         # PostgreSQL persistence and SQL migrations
│       └── types/         # Shared response types (GameStateResponse, etc.)
├── frontend/              # React single-page application (SPA)
│   ├── src/               # React components, hooks, assets, and utilities
│   └── nginx.conf         # Nginx server configuration for Docker deployment
├── .infra/
│   └── helm/              # Helm chart for Kubernetes deployment (via Kind)
├── docker-compose.yml     # Configuration for Docker services (App, DB, Redis)
└── Makefile               # Helper commands for local development and k8s
```

---

## ⚙️ Environment Configuration

To run the project locally, copy the default environment template file:

```bash
cp .env.example .env
```

The `.env` file contains settings for the database, Redis cache, CORS configurations, and secrets:
- `PORT` — Server listener port (defaults to `8080`).
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` — Postgres connection details.
- `REDIS_ADDR` — Redis connection address (defaults to `localhost:6379`).
- `COOKIE_SECRET` — Key used to sign session cookies.
- `ALLOWED_ORIGINS` — Comma-separated list of permitted origins to prevent unauthorized cross-origin requests.

---

## 🚀 Running the Application

All developer utilities are defined inside the Makefile

### Option 1: Run with Docker Compose (Recommended)
This command automatically spins up PostgreSQL, Redis, the Go API, and the React frontend in containerized environments:

```bash
# Build and run the services
make docker-up

# Stop all services
make docker-down

# View container logs
make docker-logs
```

Once running:
- Frontend client will be served at: `http://localhost:8090`
- Backend API server will be listening at: `http://localhost:8080`

### Option 2: Local Development
Ensure Go, Node.js, PostgreSQL, and Redis are installed and running on your host machine.

1. **Install frontend dependencies**:
   ```bash
   make frontend-install
   ```

2. **Start the dev servers (both frontend & backend concurrently)**:
   ```bash
   make dev
   ```
   *Alternatively, you can run them individually:*
   - Start backend: `make dev-backend`
   - Start frontend: `make frontend-dev`

---

## 🧪 Testing & Linting

Verify backend codebase correctness using:

```bash
# Run unit tests
make test

# Run golangci-lint
make lint
```

---

## 🤖 Playing with AI Bots (Reinforcement Learning)

Lil Poker is fully compatible with the reinforcement learning bot trained in the [lil-poker-rl](https://github.com/LilRaime/lil-poker-rl) repository. You can play against the bot either inside Docker containers or by running the bot agent script locally.

### Prerequisites

You must first build the bot's Docker image locally:
```bash
cd ../lil-poker-rl # Navigate to the bot repository
docker build -t lil-poker-rl .
```

### Option A: Spawning Bots via UI (Docker)

If you are running Lil Poker via Docker Compose (`make docker-up`), the Go API mounts the local `/var/run/docker.sock` to dynamically spawn bot containers on demand.

1. Open `http://localhost:8090` in your browser.
2. Create or join a room.
3. Click the **Add Bot** button inside the lobby panel.
4. A new container named after the bot (e.g., `RL_Bot_453`) will be spun up.
5. If the bot is kicked (by clicking the **×** button on its player card in the UI) or stands up, its container will cleanly exit and automatically delete itself.

### Option B: Running Bots Locally (without Docker)

You can run the bot script on your host machine to connect to a running room.

1. Navigate to the `lil-poker-rl` directory on your host:
   ```bash
   cd ../lil-poker-rl
   ```
2. Activate the virtual environment:
   ```bash
   source .venv/bin/activate
   ```
3. Find your game room ID from the browser URL (e.g., `EN3FEM` from `http://localhost:8090/?room=EN3FEM`).
4. Start the bot agent:
   ```bash
   python -m agent.play_live --room <ROOM_ID> --url http://localhost:8090 --name MyLocalBot --algo ppo --device cpu
   ```

---

## 📄 License

This project is licensed under the **GNU General Public License v3 (GPL-3.0)**