import { useState } from "react";
import { WinnerStatus } from "../types";
import Card from "./Card";

interface WinnersOverlayProps {
  lastWinners: WinnerStatus[];
  onStartGame: () => void;
  onResetGame: () => void;
  isCreator: boolean;
}

export default function WinnersOverlay({
  lastWinners,
  onStartGame,
  onResetGame,
  isCreator,
}: WinnersOverlayProps) {
  const [isMinimized, setIsMinimized] = useState(true);

  if (isMinimized) {
    return (
      <div className="fixed bottom-16 sm:bottom-6 left-3 right-3 sm:left-6 sm:right-auto sm:w-80 z-50 glass-panel-heavy p-3 sm:p-4 rounded-2xl shadow-2xl border border-amber-500/30 flex flex-col space-y-2 sm:space-y-3 animate-fade-in overflow-hidden">
        <div className="flex justify-between items-center border-b border-white/5 pb-2">
          <span className="flex items-center space-x-2 text-amber-300 font-black text-xs uppercase tracking-wider">
            🏆 Hand Winner
          </span>
          <button
            onClick={() => setIsMinimized(false)}
            className="text-xxxxs sm:text-xxs px-2.5 py-1 rounded-lg bg-amber-500/20 hover:bg-amber-500/35 text-amber-300 transition-colors font-bold uppercase tracking-wider"
          >
            🔍 Maximize
          </button>
        </div>

        <div className="flex flex-col space-y-2">
          {lastWinners.map((winner, idx) => (
            <div key={idx} className="text-xs text-slate-200 font-extrabold flex flex-col border-b border-white/5 last:border-b-0 pb-1.5 last:pb-0">
              <div className="flex justify-between items-center">
                <span>{winner.player_name}</span>
                <span className="text-amber-400 font-mono font-black">+{winner.amount} 🪙</span>
              </div>
              {winner.hand_rank && (
                <span className="text-[10px] text-indigo-300 font-semibold mt-0.5">
                  Hand: {winner.hand_rank}
                </span>
              )}
            </div>
          ))}
        </div>

        {isCreator ? (
          <button
            onClick={onStartGame}
            className="w-full py-2 bg-amber-500 hover:bg-amber-400 text-slate-950 font-black rounded-lg text-xs shadow-md transition-colors uppercase tracking-wider"
          >
            Start Next Deal
          </button>
        ) : (
          <div className="w-full py-2 px-3 bg-slate-950/60 rounded-lg border border-white/5 flex items-center justify-center gap-1.5">
            <span className="animate-pulse text-amber-400 text-[8px]">●</span>
            <span className="text-[10px] text-slate-400 font-semibold">Waiting for next deal...</span>
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-sm">
      <div className="w-full max-w-md bg-slate-900/95 border border-white/10 rounded-2xl shadow-2xl p-4 sm:p-6 flex flex-col items-center relative overflow-hidden animate-fade-in">
        <button
          onClick={() => setIsMinimized(true)}
          className="absolute top-4 right-4 text-[10px] font-extrabold text-slate-300 hover:text-slate-100 transition-colors bg-white/5 hover:bg-white/10 px-2.5 py-1.5 rounded-lg border border-white/10 z-10"
          title="Minimize overlay to see opponents' cards on the table"
        >
          👁️ View Table
        </button>

        <div className="absolute -top-12 -right-12 w-32 h-32 bg-amber-500/10 blur-2xl rounded-full pointer-events-none" />
        <div className="text-3xl mb-2">🏆</div>
        <h3 className="text-lg font-black tracking-wide text-amber-300">
          Hand Winner!
        </h3>

        <div className="w-full space-y-3 mt-4">
          {lastWinners.map((winner, idx) => (
            <div
              key={idx}
              className="bg-slate-950 border border-white/5 rounded-xl p-4 flex flex-col items-center"
            >
              <span className="font-extrabold text-slate-200 text-center">
                {winner.player_name} won (+{winner.amount} 🪙)
              </span>

              {winner.hand_rank && (
                <span className="text-xs text-indigo-300 font-semibold mt-1">
                  Hand: {winner.hand_rank}
                </span>
              )}

              {winner.hand_cards && winner.hand_cards.length > 0 && (
                <div className="flex space-x-1 mt-3">
                  {winner.hand_cards.map((card, cardIdx) => (
                    <Card key={cardIdx} cardStr={card} />
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>

        {isCreator ? (
          <div className="flex gap-2 w-full mt-5">
            <button
              onClick={onStartGame}
              className="flex-1 py-3 bg-amber-500 hover:bg-amber-400 text-slate-950 font-black rounded-xl text-sm shadow-lg transition-colors"
            >
              Start New Deal
            </button>
            <button
              onClick={onResetGame}
              className="flex-1 py-3 bg-slate-800 hover:bg-slate-700 font-bold rounded-xl text-sm border border-white/5 transition-colors"
            >
              Reset Table
            </button>
          </div>
        ) : (
          <div className="w-full mt-5 py-3.5 px-4 bg-slate-950/60 rounded-xl border border-white/5 flex items-center justify-center gap-2">
            <span className="animate-pulse text-amber-400 text-base">●</span>
            <span className="text-xs text-slate-400 font-semibold">Waiting for host to start next deal...</span>
          </div>
        )}
      </div>
    </div>
  );
}
