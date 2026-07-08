import { useEffect, useState } from "react";
import { GameStateResponse, PlayerStatus } from "../types";
import { ConnectionState } from "../hooks/usePokerSocket";

interface HeaderProps {
  gameState: GameStateResponse | null;
  isAuthenticated: boolean;
  playerName: string;
  hero: PlayerStatus | undefined;
  connectionState: ConnectionState;
  onSitToggle: () => void;
  onLeave: () => void;
  autoRebuy: boolean;
  setAutoRebuy: (val: boolean) => void;
  currentRoomId: string | null;
  onLeaveRoom: () => void;
  onStandUp: () => void;
}

export default function Header({
  gameState,
  isAuthenticated,
  playerName,
  hero,
  connectionState,
  onSitToggle,
  onLeave,
  autoRebuy,
  setAutoRebuy,
  currentRoomId,
  onLeaveRoom,
  onStandUp,
}: HeaderProps) {
  const [isLight, setIsLight] = useState(() => localStorage.getItem("theme") === "light");
  const [timeUntilRaise, setTimeUntilRaise] = useState<string | null>(null);

  useEffect(() => {
    if (isLight) {
      document.body.classList.add("light-theme");
    } else {
      document.body.classList.remove("light-theme");
    }
  }, [isLight]);

  const toggleTheme = () => {
    setIsLight((prev) => {
      const next = !prev;
      localStorage.setItem("theme", next ? "light" : "dark");
      return next;
    });
  };

  const blindsRaiseDeadline = gameState?.blinds_raise_deadline;
  const phase = gameState?.phase;

  useEffect(() => {
    if (!blindsRaiseDeadline || phase === "Waiting") {
      setTimeUntilRaise(null);
      return;
    }

    const updateTimer = () => {
      const remainingMs = blindsRaiseDeadline - Date.now();
      if (remainingMs <= 0) {
        setTimeUntilRaise("00:00");
        return;
      }
      const totalSeconds = Math.floor(remainingMs / 1000);
      const minutes = Math.floor(totalSeconds / 60);
      const seconds = totalSeconds % 60;
      setTimeUntilRaise(
        `${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`
      );
    };

    updateTimer();
    const interval = setInterval(updateTimer, 1000);
    return () => clearInterval(interval);
  }, [blindsRaiseDeadline, phase]);

  return (
    <header className="glass-panel-heavy border-b border-white/5 py-1.5 px-3 sm:py-2.5 sm:px-6 flex justify-between items-center z-10 sticky top-0">
      <div className="hidden sm:flex items-center space-x-1.5 sm:space-x-3">
        <span className="text-xl sm:text-2xl text-purple-400">♠</span>
        <div>
          <div className="flex items-center gap-1.5 sm:gap-2">
            <h1 className="text-sm sm:text-lg font-extrabold tracking-wider bg-clip-text text-transparent bg-gradient-to-r from-purple-400 to-indigo-400">
              LIL-POKER
            </h1>
            <span
              className={`w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full inline-block ${connectionState === "connected"
                ? "bg-emerald-500 shadow-[0_0_8px_#10b981] animate-pulse"
                : connectionState === "connecting"
                  ? "bg-amber-500 shadow-[0_0_8px_#f59e0b] animate-pulse"
                  : "bg-red-500 shadow-[0_0_8px_#ef4444]"
                }`}
              title={`WebSocket: ${connectionState}`}
            />
          </div>
          <p className="hidden sm:block text-xxs text-slate-400 uppercase tracking-widest leading-none">
            Real-time Engine
          </p>
        </div>
      </div>

      {currentRoomId && (
        <div className="flex items-center space-x-3.5 sm:space-x-6">
          {gameState && (
            <>
              <div className="text-center select-none">
                <div className="text-xxs uppercase tracking-widest text-slate-500 font-bold leading-none mb-1">Phase</div>
                <div className="font-bold text-indigo-300 leading-tight text-xs sm:text-sm whitespace-nowrap">{gameState.phase}</div>
                {gameState.observer_count !== undefined && gameState.observer_count > 0 && (
                  <div className="hidden sm:block text-[9px] text-slate-500 font-bold leading-none mt-1">
                    👁️ {gameState.observer_count} observing
                  </div>
                )}
              </div>
              <div className="hidden sm:block h-8 w-px bg-white/10" />
              <div className="text-center select-none">
                <div className="text-xxs uppercase tracking-widest text-slate-500 font-bold leading-none mb-1">Mode</div>
                {gameState.starting_chips > 0 ? (
                  <div className="text-xs sm:text-sm font-black text-purple-300 leading-tight whitespace-nowrap">🏆 Tournament</div>
                ) : (
                  <div className="text-xs sm:text-sm font-black text-emerald-300 leading-tight whitespace-nowrap">💰 Persistent</div>
                )}
              </div>
              <div className="hidden sm:block h-8 w-px bg-white/10" />
              <div className="select-none">
                <div className="flex items-center gap-2">
                  <div>
                    <div className="text-xxs uppercase tracking-widest text-slate-500 font-bold leading-none mb-1">Blind</div>
                    <div className="font-black text-slate-200 text-xs sm:text-sm leading-none whitespace-nowrap">{gameState.small_blind}/{gameState.big_blind}</div>
                  </div>
                  {gameState.phase !== "Waiting" && gameState.blinds_raise_deadline > 0 && gameState.next_small_blind !== undefined && (
                    <div className="text-center">
                      <div className="text-xxs text-slate-500 font-bold leading-none mb-1 whitespace-nowrap">{timeUntilRaise ?? ""}</div>
                      <div className="font-black text-slate-400 text-xs sm:text-sm leading-none whitespace-nowrap">
                        {gameState.next_small_blind === gameState.small_blind ? (
                          <span className="text-amber-500/75">cap</span>
                        ) : (
                          <span>{gameState.next_small_blind}/{gameState.next_big_blind}</span>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              </div>
              <div className="hidden sm:block h-8 w-px bg-white/10" />
            </>
          )}

          <div className="text-center select-none">
            <div className="text-xxs uppercase tracking-widest text-slate-500 font-bold leading-none mb-1">Room</div>
            <div className="font-mono font-black text-slate-200 text-xs sm:text-sm tracking-widest leading-none">{currentRoomId}</div>
          </div>

          <div className="hidden sm:block h-8 w-px bg-white/10" />

          <div className="hidden sm:block text-center select-none">
            <div className="text-xxs uppercase tracking-widest text-slate-500 font-bold leading-none mb-1">Exit</div>
            <button
              onClick={onLeaveRoom}
              className="text-red-400 hover:text-red-300 hover:underline text-xs sm:text-sm font-black leading-none"
            >
              Leave
            </button>
          </div>
        </div>
      )}

      {isAuthenticated && (
        <div className="flex items-center gap-1.5 sm:gap-4">
          <button
            onClick={toggleTheme}
            className="w-7 h-7 sm:w-8 sm:h-8 rounded-lg flex items-center justify-center bg-slate-800 border border-white/5 hover:bg-slate-700 transition-colors text-slate-300 text-xs sm:text-sm shadow-md"
            title="Toggle Light/Dark Mode"
          >
            {isLight ? "🌙" : "☀️"}
          </button>
          <div className="text-right">
            <div className="text-[10px] sm:text-xs font-semibold text-slate-300 leading-none">{playerName}</div>
            {hero && <div className="text-[10px] sm:text-xs text-emerald-400 font-bold mt-0.5">{hero.chips} 🪙</div>}
          </div>
          {hero && (
            <div className="hidden sm:flex items-center gap-1 sm:gap-3">
              <label className="flex items-center space-x-1 sm:space-x-2 text-[10px] sm:text-xs font-semibold text-slate-300 cursor-pointer select-none bg-slate-800 border border-white/5 hover:bg-slate-700 px-1.5 py-1 sm:px-3 sm:py-1.5 rounded-lg transition-colors">
                <input
                  type="checkbox"
                  checked={autoRebuy}
                  onChange={(e) => setAutoRebuy(e.target.checked)}
                  className="accent-purple-500 rounded"
                />
                <span>Auto<span className="hidden sm:inline">-Rebuy</span></span>
              </label>
              <button
                onClick={onSitToggle}
                className={`px-2 py-1 sm:px-3 sm:py-1.5 rounded-lg text-[10px] sm:text-xs font-semibold border transition-all ${hero.sitting_out
                  ? "bg-amber-600/20 text-amber-300 border-amber-500/20 hover:bg-amber-600/30"
                  : "bg-slate-800 text-slate-300 border-white/5 hover:bg-slate-700"
                  }`}
              >
                {hero.sitting_out ? "Sit In" : "Sit Out"}
              </button>
              <button
                onClick={onStandUp}
                className="px-2 py-1 sm:px-3 sm:py-1.5 rounded-lg text-[10px] sm:text-xs font-semibold border bg-red-950/20 hover:bg-red-950/40 text-red-400 border-red-500/10 hover:border-red-500/20 transition-all cursor-pointer"
              >
                Stand Up
              </button>
            </div>
          )}
          <button
            onClick={onLeave}
            className="hidden sm:block px-2 py-1 sm:px-3.5 sm:py-1.5 rounded-lg text-[10px] sm:text-xs font-semibold text-red-400 hover:text-red-300 bg-red-950/20 hover:bg-red-950/40 border border-red-500/10 transition-colors"
          >
            Logout
          </button>
        </div>
      )}
    </header>
  );
}
