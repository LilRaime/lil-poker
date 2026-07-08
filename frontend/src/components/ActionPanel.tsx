import { useEffect, useState, useCallback } from "react";
import { GameStateResponse } from "../types";

interface ActionPanelProps {
  gameState: GameStateResponse;
  playerId: string;
  onAction: (action: string, amount?: number) => void;
}

export default function ActionPanel({ gameState, playerId, onAction }: ActionPanelProps) {
  const hero = gameState.players.find((p) => p.id === playerId);
  const isHeroActive =
    gameState.active_player_id
      ? gameState.active_player_id === playerId
      : gameState.players[gameState.active_idx]?.id === playerId;

  const currentBet = gameState.current_bet;
  const heroBet = hero?.bet || 0;
  const toCall = currentBet - heroBet;
  const heroChips = hero?.chips || 0;

  const minRaise = currentBet + gameState.big_blind;
  const maxRaise = heroChips + heroBet;
  const clampRaise = (val: number) => Math.min(maxRaise, Math.max(minRaise, val));

  const [raiseAmount, setRaiseAmount] = useState(minRaise);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    setRaiseAmount(Math.min(minRaise, maxRaise));
  }, [currentBet, minRaise, maxRaise]);

  const handleButtonClick = useCallback(async (action: string, amount?: number) => {
    if (submitting) return;
    setSubmitting(true);
    try {
      await onAction(action, amount);
    } finally {
      setSubmitting(false);
    }
  }, [submitting, onAction]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (!isHeroActive || submitting) return;
      const tag = (e.target as HTMLElement).tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      switch (e.key.toLowerCase()) {
        case "f":
          handleButtonClick("fold");
          break;
        case "c":
          handleButtonClick(toCall === 0 ? "check" : "call");
          break;
        case "r":
          if (heroChips > toCall && raiseAmount >= minRaise) {
            handleButtonClick("raise", raiseAmount);
          }
          break;
        case "a":
          handleButtonClick("all_in");
          break;
      }
    },
    [isHeroActive, submitting, toCall, heroChips, raiseAmount, minRaise, handleButtonClick]
  );

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);

  if (!isHeroActive) {
    return (
      <div className="bg-slate-900/40 border border-white/5 py-4 rounded-xl text-center text-sm font-semibold text-slate-400 italic">
        Waiting for opponent action...
      </div>
    );
  }

  const btnClass = "h-full py-4 px-3 sm:px-5 font-black rounded-xl text-sm sm:text-base transition-all disabled:opacity-50 disabled:cursor-not-allowed select-none active:scale-[0.98] border hover:shadow-lg flex items-center justify-center whitespace-nowrap";

  return (
    <div className="flex flex-col space-y-4">
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
        <button
          disabled={submitting}
          onClick={() => handleButtonClick("fold")}
          className={`${btnClass} bg-red-950/40 hover:bg-red-950/60 border border-red-500/30 text-red-200 hover:text-red-100`}
        >
          Fold <span className="hidden sm:inline text-red-400/50 text-xs font-normal ml-1">[F]</span>
        </button>

        {toCall === 0 ? (
          <button
            disabled={submitting}
            onClick={() => handleButtonClick("check")}
            className={`${btnClass} bg-blue-950/40 hover:bg-blue-950/60 border border-blue-500/30 text-blue-200 hover:text-blue-100`}
          >
            Check <span className="hidden sm:inline text-blue-400/50 text-xs font-normal ml-1">[C]</span>
          </button>
        ) : (
          <button
            disabled={submitting}
            onClick={() => handleButtonClick("call")}
            className={`${btnClass} bg-emerald-950/40 hover:bg-emerald-950/60 border border-emerald-500/30 text-emerald-200 hover:text-emerald-100 flex flex-col items-center justify-center`}
          >
            <span>Call ({toCall}) <span className="hidden sm:inline text-emerald-400/50 text-xs font-normal ml-1">[C]</span></span>
            <span className="text-[10px] text-emerald-400/60 font-semibold leading-none mt-0.5">
              Odds: {Math.round((toCall / (gameState.pot + toCall)) * 100)}% pot
            </span>
          </button>
        )}

        {heroChips > toCall && (
          <button
            disabled={submitting}
            onClick={() => handleButtonClick("raise", raiseAmount)}
            className={`${btnClass} bg-purple-900/50 hover:bg-purple-900/70 border border-purple-500/30 text-purple-200 hover:text-purple-100`}
          >
            Raise ({raiseAmount}) <span className="hidden sm:inline text-purple-400/50 text-xs font-normal ml-1">[R]</span>
          </button>
        )}

        <button
          disabled={submitting}
          onClick={() => handleButtonClick("all_in")}
          className={`${btnClass} bg-amber-900/40 hover:bg-amber-900/60 border border-amber-500/30 text-amber-200 hover:text-amber-100`}
        >
          All-In <span className="hidden sm:inline text-amber-400/50 text-xs font-normal ml-1">[A]</span>
        </button>
      </div>

      {heroChips > toCall && minRaise <= maxRaise && (
        <div className="bg-slate-900/80 border border-white/5 rounded-xl p-3 flex flex-col space-y-2">
          <div className="flex justify-between items-center text-2xs text-slate-400 font-bold uppercase tracking-wider">
            <span>Min: {minRaise}</span>
            <span className="text-purple-300 font-extrabold text-xs">
              Raise Amount: {raiseAmount}
            </span>
            <span>Max: {maxRaise}</span>
          </div>
          <input
            disabled={submitting}
            type="range"
            min={minRaise}
            max={maxRaise}
            step={5}
            value={raiseAmount}
            onChange={(e) => setRaiseAmount(Number(e.target.value))}
            className="w-full accent-purple-500 disabled:opacity-50"
          />
          <div className="flex justify-between gap-1.5 mt-2">
            {[
              { label: "Min", val: minRaise },
              { label: "1/2 Pot", val: clampRaise(Math.floor(gameState.pot / 2)) },
              { label: "Pot", val: clampRaise(gameState.pot) },
              { label: "Max", val: maxRaise }
            ].map(({ label, val }) => (
              <button
                key={label}
                disabled={submitting}
                type="button"
                onClick={() => setRaiseAmount(val)}
                className="flex-1 py-2 sm:py-2.5 px-2 sm:px-3 text-[10px] sm:text-xs font-black rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 border border-white/10 transition-colors shadow-sm"
              >
                {label} ({val})
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
