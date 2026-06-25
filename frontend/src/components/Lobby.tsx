import React from "react";

interface LobbyProps {
  playerName: string;
  playerChips: number;
  onJoin: (e: React.FormEvent) => void;
  onReset: () => void;
  onLogout: () => void;
  onRebuy: () => void;
}

export default function Lobby({
  playerName,
  playerChips,
  onJoin,
  onReset,
  onLogout,
  onRebuy,
}: LobbyProps) {
  return (
    <div className="w-full max-w-md glass-panel p-5 sm:p-8 rounded-2xl shadow-2xl relative overflow-hidden border border-white/10 text-center animate-fade-in">
      <div className="absolute -top-12 -right-12 w-36 h-36 bg-purple-600/10 blur-3xl rounded-full" />
      <div className="absolute -bottom-12 -left-12 w-36 h-36 bg-indigo-600/10 blur-3xl rounded-full" />

      <div className="mb-6">
        <div className="text-5xl mb-4">👑</div>
        <h2 className="text-2xl font-black tracking-wide text-slate-100">Welcome, {playerName}!</h2>
        <p className="text-indigo-400 font-extrabold text-sm mt-2 flex items-center justify-center">
          Balance: <span className="text-amber-400 ml-1.5 text-base">{playerChips} 🪙</span>
        </p>
      </div>

      <form onSubmit={onJoin} className="space-y-4 mt-6">
        {playerChips < 10 ? (
          <button
            type="button"
            onClick={onRebuy}
            className="w-full bg-gradient-to-r from-amber-600 to-yellow-600 hover:from-amber-500 hover:to-yellow-500 py-3.5 rounded-xl font-bold tracking-wide shadow-lg hover:shadow-yellow-500/20 active:translate-y-px transition-all text-white"
          >
            Claim Rebuy (1,000 🪙)
          </button>
        ) : (
          <button
            type="submit"
            className="w-full bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 py-3.5 rounded-xl font-bold tracking-wide shadow-lg hover:shadow-indigo-500/20 active:translate-y-px transition-all text-white"
          >
            Take a Seat
          </button>
        )}

        <div className="grid grid-cols-2 gap-3 mt-2">
          <button
            type="button"
            onClick={onReset}
            className="bg-slate-900 hover:bg-slate-800 text-slate-400 hover:text-slate-300 py-2.5 rounded-xl font-bold text-xs border border-white/5 transition-all"
          >
            Reset Table
          </button>
          <button
            type="button"
            onClick={onLogout}
            className="bg-red-950/20 hover:bg-red-950/40 text-red-400 hover:text-red-300 py-2.5 rounded-xl font-bold text-xs border border-red-500/10 transition-all"
          >
            Logout
          </button>
        </div>
      </form>
    </div>
  );
}
