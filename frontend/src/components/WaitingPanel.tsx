import { useState } from "react";
import { GameStateResponse } from "../types";
import { BLIND_PRESETS } from "../constants";

interface WaitingPanelProps {
  gameState: GameStateResponse;
  isCreator: boolean;
  onStart: () => void;
  onReset: () => void;
  onSetBlinds: (sb: number, bb: number) => void;
  onAddBot: () => void;
}

export default function WaitingPanel({
  gameState,
  isCreator,
  onStart,
  onReset,
  onSetBlinds,
  onAddBot,
}: WaitingPanelProps) {
  const canStart = isCreator && gameState.players.length >= 2;
  const [collapsed, setCollapsed] = useState(true);

  return (
    <div className="mb-3 sm:mb-6 w-full max-w-2xl glass-panel rounded-2xl border border-white/5 shadow-xl overflow-hidden">
      {/* ── Mobile compact header ── */}
      <div
        className="sm:hidden flex items-center justify-between px-4 py-2.5 cursor-pointer select-none active:bg-white/5 transition-colors"
        onClick={() => setCollapsed(!collapsed)}
      >
        <div className="flex items-center gap-2">
          <span className="text-[10px] uppercase tracking-widest text-slate-500 font-bold">⏳ Waiting</span>
          <span className="text-xs font-bold text-slate-300">
            · {gameState.players.length} player{gameState.players.length !== 1 ? "s" : ""}
          </span>
          {!isCreator && (
            <span className="text-[10px] text-amber-400/80 font-semibold">· host controls</span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {canStart && (
            <button
              onClick={(e) => { e.stopPropagation(); onStart(); }}
              className="px-3 py-1 bg-indigo-600 hover:bg-indigo-500 font-bold rounded-lg text-xs transition-colors"
            >
              Start
            </button>
          )}
          <span className="text-slate-500 text-xs transition-transform duration-200" style={{ display: "inline-block", transform: collapsed ? "rotate(0deg)" : "rotate(180deg)" }}>▼</span>
        </div>
      </div>

      <div className={`${collapsed ? "hidden" : "block"} sm:block px-4 sm:px-6 py-4 sm:py-5`}>
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="text-left">
            <h3 className="font-bold text-slate-200 text-sm sm:text-base">Waiting to Start Game</h3>
            <p className="text-xs text-slate-400 mt-0.5">
              At least 2 players required. Current players: {gameState.players.length}
            </p>

            {!isCreator && (
              <p className="text-xs text-amber-400/80 mt-1 flex items-center gap-1">
                <span>⏳</span> Waiting for the host to start the game…
              </p>
            )}

            <div className="mt-3 flex flex-wrap items-center gap-2">
              <span className="text-xs uppercase tracking-wider text-slate-400 font-bold">Blinds:</span>
              <div className="flex flex-wrap gap-1.5">
                {BLIND_PRESETS.map(([sb, bb]) => {
                  const active = gameState.small_blind === sb && gameState.big_blind === bb;
                  return (
                    <button
                      key={`${sb}-${bb}`}
                      onClick={() => isCreator && onSetBlinds(sb, bb)}
                      disabled={!isCreator}
                      className={`px-2.5 py-1 text-xs font-bold rounded-lg border transition-all ${active
                          ? "bg-indigo-600/40 text-indigo-200 border-indigo-500"
                          : isCreator
                            ? "bg-slate-800 border-white/10 text-slate-300 hover:bg-slate-700 hover:text-white"
                            : "bg-slate-800/40 text-slate-600 border-white/5 cursor-not-allowed"
                        }`}
                    >
                      {sb}/{bb}
                    </button>
                  );
                })}
              </div>
            </div>
          </div>

          <div className="flex gap-2 w-full sm:w-auto">
            {isCreator ? (
              <>
                {gameState.players.length >= 2 && (
                  <button
                    onClick={onStart}
                    disabled={!canStart}
                    className="flex-1 sm:flex-none px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 font-bold rounded-xl text-sm shadow transition-colors"
                  >
                    Start<span className="hidden sm:inline"> Deal</span>
                  </button>
                )}
                <button
                  onClick={onReset}
                  className="flex-1 sm:flex-none px-4 py-2.5 bg-slate-800 hover:bg-slate-700 font-bold rounded-xl text-sm border border-white/5 transition-colors"
                >
                  Reset<span className="hidden sm:inline"> Table</span>
                </button>
                <button
                  onClick={onAddBot}
                  disabled={gameState.players.length >= 8}
                  className="flex-1 sm:flex-none px-4 py-2.5 bg-purple-950/50 hover:bg-purple-900/60 disabled:opacity-50 text-purple-300 border border-purple-500/20 hover:scale-105 active:scale-95 font-bold rounded-xl text-sm transition-all shadow-md"
                >
                  Add Bot
                </button>
              </>
            ) : (
              <div className="flex items-center gap-2 px-4 py-2.5 bg-slate-800/40 rounded-xl border border-white/5">
                <span className="animate-pulse text-indigo-400 text-base">●</span>
                <span className="text-xs text-slate-400 font-semibold">Host controls game</span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
