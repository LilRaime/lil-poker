import { GameStateResponse } from "../types";
import { BLIND_PRESETS } from "../constants";

interface WaitingPanelProps {
  gameState: GameStateResponse;
  isCreator: boolean;
  onStart: () => void;
  onReset: () => void;
  onSetBlinds: (sb: number, bb: number) => void;
}

export default function WaitingPanel({
  gameState,
  isCreator,
  onStart,
  onReset,
  onSetBlinds,
}: WaitingPanelProps) {
  const canStart = isCreator && gameState.players.length >= 2;

  return (
    <div className="mb-6 w-full max-w-2xl text-center glass-panel px-6 py-5 rounded-2xl flex flex-col sm:flex-row items-center justify-between gap-4 border border-white/5 shadow-xl">
      <div className="text-left">
        <h3 className="font-bold text-slate-200">Waiting to Start Game</h3>
        <p className="text-xs text-slate-400 mt-0.5">
          At least 2 players required. Current players: {gameState.players.length}
        </p>

        {!isCreator && (
          <p className="text-xs text-amber-400/80 mt-1 flex items-center gap-1">
            <span>⏳</span> Waiting for the host to start the game…
          </p>
        )}

        <div className="mt-3 flex items-center space-x-2">
          <span className="text-xs uppercase tracking-wider text-slate-400 font-bold">Blinds:</span>
          <div className="flex gap-1.5">
            {BLIND_PRESETS.map(([sb, bb]) => {
              const active =
                gameState.small_blind === sb && gameState.big_blind === bb;
              return (
                <button
                  key={`${sb}-${bb}`}
                  onClick={() => isCreator && onSetBlinds(sb, bb)}
                  disabled={!isCreator}
                  className={`px-3 py-1.5 text-xs font-bold rounded-lg border transition-all ${
                    active
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
                Start Deal
              </button>
            )}
            <button
              onClick={onReset}
              className="flex-1 sm:flex-none px-4 py-2.5 bg-slate-800 hover:bg-slate-700 font-bold rounded-xl text-sm border border-white/5 transition-colors"
            >
              Reset Table
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
  );
}
