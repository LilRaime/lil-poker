const PLAYER_COLORS: Array<{ border: string; text: string; bg: string; dot: string }> = [
  { border: "border-indigo-500", text: "text-indigo-400", bg: "bg-indigo-500/20", dot: "bg-indigo-500" },
  { border: "border-emerald-500", text: "text-emerald-400", bg: "bg-emerald-500/20", dot: "bg-emerald-500" },
  { border: "border-amber-500", text: "text-amber-400", bg: "bg-amber-500/20", dot: "bg-amber-500" },
  { border: "border-rose-500", text: "text-rose-400", bg: "bg-rose-500/20", dot: "bg-rose-500" },
  { border: "border-cyan-500", text: "text-cyan-400", bg: "bg-cyan-500/20", dot: "bg-cyan-500" },
  { border: "border-violet-500", text: "text-violet-400", bg: "bg-violet-500/20", dot: "bg-violet-500" },
  { border: "border-orange-500", text: "text-orange-400", bg: "bg-orange-500/20", dot: "bg-orange-500" },
  { border: "border-teal-500", text: "text-teal-400", bg: "bg-teal-500/20", dot: "bg-teal-500" },
];

export function getPlayerColor(playerId: string) {
  let hash = 0;
  for (let i = 0; i < Math.min(playerId.length, 8); i++) {
    hash = (hash * 31 + playerId.charCodeAt(i)) & 0xff;
  }
  return PLAYER_COLORS[hash % PLAYER_COLORS.length];
}
