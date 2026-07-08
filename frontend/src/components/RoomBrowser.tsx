import { useState, useEffect, useCallback } from "react";
import { RoomInfo } from "../types";
import { pokerApi } from "../services/api";

interface RoomBrowserProps {
  playerName: string;
  playerChips: number;
  onJoinRoom: (roomId: string, creatorId?: string) => void;
  onLogout: () => void;
}

export default function RoomBrowser({ playerName, playerChips, onJoinRoom, onLogout }: RoomBrowserProps) {
  const [rooms, setRooms] = useState<RoomInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const [showCreate, setShowCreate] = useState(false);
  const [roomName, setRoomName] = useState("");
  const [smallBlind, setSmallBlind] = useState(10);
  const [bigBlind, setBigBlind] = useState(20);
  const [creating, setCreating] = useState(false);
  const [blindEscalationMins, setBlindEscalationMins] = useState(5);
  const [chipMode, setChipMode] = useState<"tournament" | "persistent">("tournament");
  const [startingChips, setStartingChips] = useState(1000);

  const [joinCode, setJoinCode] = useState("");
  const [showJoinCode, setShowJoinCode] = useState(false);

  const fetchRooms = useCallback(async () => {
    try {
      const data = await pokerApi.listRooms();
      setRooms(data);
    } catch {
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchRooms();
    const interval = setInterval(fetchRooms, 4000);
    return () => clearInterval(interval);
  }, [fetchRooms]);

  const handleCreateRoom = async (e: React.FormEvent) => {
    e.preventDefault();
    if (creating) return;
    if (bigBlind < smallBlind) {
      setErrorMsg("Big blind must be greater than or equal to small blind");
      return;
    }
    setCreating(true);
    setErrorMsg(null);
    try {
      const room = await pokerApi.createRoom(
        roomName || `${playerName}'s Table`,
        smallBlind,
        bigBlind,
        blindEscalationMins,
        chipMode === "tournament" ? startingChips : 0
      );
      onJoinRoom(room.id, room.creator_id);
    } catch (err: any) {
      setErrorMsg(err.message || "Failed to create room");
      setCreating(false);
    }
  };

  const handleJoinByCode = (e: React.FormEvent) => {
    e.preventDefault();
    const code = joinCode.trim().toUpperCase();
    if (!code) return;
    onJoinRoom(code);
  };

  const phaseLabel = (phase: string) => {
    if (phase === "Waiting") return { text: "Waiting", cls: "bg-slate-700/60 text-slate-400 border-slate-600/30" };
    return { text: "In Progress", cls: "bg-emerald-500/20 text-emerald-300 border-emerald-500/30" };
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-4">
      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        <div className="absolute top-1/4 left-1/2 -translate-x-1/2 w-[600px] h-[600px] rounded-full bg-purple-600/5 blur-3xl" />
      </div>
      <div className="w-full max-w-3xl flex items-center justify-between mb-8">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-600 to-indigo-700 flex items-center justify-center text-xl shadow-lg shadow-purple-900/40">
            ♠
          </div>
          <div>
            <div className="text-slate-100 font-black text-lg leading-none">lil-poker</div>
            <div className="text-slate-500 text-xs">Room Lobby</div>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <div className="text-right">
            <div className="text-slate-300 text-sm font-bold">{playerName}</div>
            <div className="text-amber-400 text-xs font-mono">
              {playerChips.toLocaleString()} chips
            </div>
          </div>
          <button
            onClick={onLogout}
            className="text-slate-500 hover:text-slate-300 text-xs px-3 py-1.5 rounded-lg border border-white/5 hover:border-white/10 transition-all"
          >
            Logout
          </button>
        </div>
      </div>

      <div className="w-full max-w-3xl space-y-4">
        <div className="flex gap-3">
          <button
            onClick={() => { setShowCreate(v => !v); setShowJoinCode(false); setErrorMsg(null); }}
            className={`flex-1 flex items-center justify-center gap-2 py-3 px-4 rounded-xl font-bold text-sm transition-all border ${showCreate
              ? "bg-purple-600/30 border-purple-500/40 text-purple-200"
              : "bg-slate-900 border-white/5 text-slate-300 hover:border-purple-500/30 hover:text-purple-300"
              }`}
          >
            <span className="text-base">+</span> Create Room
          </button>
          <button
            onClick={() => { setShowJoinCode(v => !v); setShowCreate(false); setErrorMsg(null); }}
            className={`flex-1 flex items-center justify-center gap-2 py-3 px-4 rounded-xl font-bold text-sm transition-all border ${showJoinCode
              ? "bg-indigo-600/30 border-indigo-500/40 text-indigo-200"
              : "bg-slate-900 border-white/5 text-slate-300 hover:border-indigo-500/30 hover:text-indigo-300"
              }`}
          >
            <span className="text-base">🔗</span> Join by Code
          </button>
        </div>

        {errorMsg && (
          <div className="bg-red-950/60 border border-red-500/30 text-red-300 text-sm px-4 py-3 rounded-xl flex items-center gap-2">
            <span>⚠️</span> {errorMsg}
          </div>
        )}

        {showCreate && (
          <div className="bg-slate-900/80 border border-white/8 rounded-2xl p-5 space-y-4 animate-fade-in">
            <h3 className="text-slate-200 font-black text-sm uppercase tracking-widest">New Table</h3>
            <form onSubmit={handleCreateRoom} className="space-y-4">
              <div>
                <label className="text-slate-400 text-xs font-bold uppercase tracking-wider block mb-1.5">
                  Room Name
                </label>
                <input
                  type="text"
                  value={roomName}
                  onChange={e => setRoomName(e.target.value)}
                  placeholder={`${playerName}'s Table`}
                  maxLength={40}
                  className="w-full bg-slate-800/60 border border-white/5 rounded-xl px-4 py-2.5 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-purple-500/50 transition-colors"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-slate-400 text-xs font-bold uppercase tracking-wider block mb-1.5">
                    Small Blind
                  </label>
                  <input
                    type="number"
                    value={smallBlind}
                    min={1}
                    onChange={e => setSmallBlind(Number(e.target.value))}
                    className="w-full bg-slate-800/60 border border-white/5 rounded-xl px-4 py-2.5 text-sm text-slate-100 focus:outline-none focus:border-purple-500/50 transition-colors"
                  />
                </div>
                <div>
                  <label className="text-slate-400 text-xs font-bold uppercase tracking-wider block mb-1.5">
                    Big Blind
                  </label>
                  <input
                    type="number"
                    value={bigBlind}
                    min={1}
                    onChange={e => setBigBlind(Number(e.target.value))}
                    className="w-full bg-slate-800/60 border border-white/5 rounded-xl px-4 py-2.5 text-sm text-slate-100 focus:outline-none focus:border-purple-500/50 transition-colors"
                  />
                </div>
              </div>
              <div>
                <label className="text-slate-400 text-xs font-bold uppercase tracking-wider block mb-1.5">
                  Blind Escalation
                </label>
                <select
                  value={blindEscalationMins}
                  onChange={e => setBlindEscalationMins(Number(e.target.value))}
                  className="w-full bg-slate-800/60 border border-white/5 rounded-xl px-4 py-2.5 text-sm text-slate-100 focus:outline-none focus:border-purple-500/50 transition-colors"
                >
                  <option value={5}>Every 5 Minutes (Default)</option>
                  <option value={10}>Every 10 Minutes</option>
                  <option value={15}>Every 15 Minutes</option>
                  <option value={0}>Never (Static Blinds)</option>
                </select>
              </div>
              <div>
                <label className="text-slate-400 text-xs font-bold uppercase tracking-wider block mb-1.5">
                  Chip Mode
                </label>
                <div className="grid grid-cols-2 gap-2">
                  <button
                    type="button"
                    onClick={() => setChipMode("tournament")}
                    className={`py-2.5 px-3 rounded-xl text-xs font-bold border transition-all text-left ${
                      chipMode === "tournament"
                        ? "bg-purple-600/30 border-purple-500/50 text-purple-200"
                        : "bg-slate-800/60 border-white/5 text-slate-400 hover:border-white/10"
                    }`}
                  >
                    <div className="text-base mb-0.5">🏆</div>
                    <div>Tournament</div>
                    <div className="text-[10px] opacity-60 font-normal mt-0.5">Fixed starting chips</div>
                  </button>
                  <button
                    type="button"
                    onClick={() => setChipMode("persistent")}
                    className={`py-2.5 px-3 rounded-xl text-xs font-bold border transition-all text-left ${
                      chipMode === "persistent"
                        ? "bg-emerald-600/30 border-emerald-500/50 text-emerald-200"
                        : "bg-slate-800/60 border-white/5 text-slate-400 hover:border-white/10"
                    }`}
                  >
                    <div className="text-base mb-0.5">💰</div>
                    <div>Persistent</div>
                    <div className="text-[10px] opacity-60 font-normal mt-0.5">Use account chips</div>
                  </button>
                </div>
              </div>
              {chipMode === "tournament" && (
                <div>
                  <label className="text-slate-400 text-xs font-bold uppercase tracking-wider block mb-1.5">
                    Starting Chips
                  </label>
                  <input
                    type="number"
                    value={startingChips}
                    min={100}
                    step={100}
                    onChange={e => setStartingChips(Number(e.target.value))}
                    className="w-full bg-slate-800/60 border border-white/5 rounded-xl px-4 py-2.5 text-sm text-slate-100 focus:outline-none focus:border-purple-500/50 transition-colors"
                  />
                </div>
              )}
              <button
                type="submit"
                disabled={creating}
                className="w-full bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 disabled:opacity-50 text-white font-black py-3 rounded-xl text-sm transition-all shadow-lg shadow-purple-900/30"
              >
                {creating ? "Creating..." : "Create & Join Room"}
              </button>
            </form>
          </div>
        )}

        {showJoinCode && (
          <div className="bg-slate-900/80 border border-white/8 rounded-2xl p-5 space-y-4 animate-fade-in">
            <h3 className="text-slate-200 font-black text-sm uppercase tracking-widest">Enter Room Code</h3>
            <form onSubmit={handleJoinByCode} className="flex gap-3">
              <input
                type="text"
                value={joinCode}
                onChange={e => setJoinCode(e.target.value.toUpperCase())}
                placeholder="ABCD12"
                maxLength={8}
                className="flex-1 bg-slate-800/60 border border-white/5 rounded-xl px-4 py-2.5 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500/50 transition-colors font-mono tracking-widest uppercase"
              />
              <button
                type="submit"
                disabled={!joinCode.trim()}
                className="bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white font-black px-5 py-2.5 rounded-xl text-sm transition-all"
              >
                Join
              </button>
            </form>
          </div>
        )}

        <div className="bg-slate-900/50 border border-white/5 rounded-2xl overflow-hidden">
          <div className="px-5 py-3 border-b border-white/5 flex items-center justify-between">
            <span className="text-slate-400 text-xs font-extrabold uppercase tracking-widest">
              Active Tables
            </span>
            <span className="text-slate-600 text-xs">
              {loading ? "Loading..." : `${rooms.length} table${rooms.length !== 1 ? "s" : ""}`}
            </span>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-16 text-slate-600 text-sm">
              <span className="animate-spin mr-2">⟳</span> Loading rooms...
            </div>
          ) : rooms.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center px-8">
              <div className="text-4xl mb-3 opacity-30">🃏</div>
              <div className="text-slate-400 text-sm font-bold mb-1">No active tables</div>
              <div className="text-slate-600 text-xs">Create a room to start playing</div>
            </div>
          ) : (
            <div className="divide-y divide-white/5">
              {rooms.map((room) => {
                const phase = phaseLabel(room.phase);
                return (
                  <div
                    key={room.id}
                    className="flex items-center justify-between px-5 py-4 hover:bg-white/3 transition-colors group cursor-pointer"
                    onClick={() => onJoinRoom(room.id, room.creator_id)}
                  >
                    <div className="flex items-center gap-4 min-w-0">
                      <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-emerald-600/30 to-teal-700/30 border border-emerald-500/20 flex items-center justify-center text-emerald-400 text-base flex-shrink-0">
                        ♣
                      </div>
                      <div className="min-w-0">
                        <div className="text-slate-200 font-bold text-sm truncate">{room.name}</div>
                        <div className="flex items-center gap-2 mt-0.5 flex-wrap">
                          <span className="text-slate-500 text-xs font-mono">#{room.id}</span>
                          <span className="text-slate-600 text-xs">·</span>
                          <span className="text-slate-500 text-xs">
                            Blinds {room.small_blind}/{room.big_blind}
                          </span>
                          <span className="text-slate-600 text-xs">·</span>
                          <span className="text-slate-500 text-xs">
                            {room.player_count}/{room.max_players} players
                          </span>
                          <span className="text-slate-600 text-xs">·</span>
                          {room.starting_chips > 0 ? (
                            <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-purple-500/15 text-purple-400 border border-purple-500/20">
                              🏆 Tournament
                            </span>
                          ) : (
                            <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-emerald-500/15 text-emerald-400 border border-emerald-500/20">
                              💰 Persistent
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 flex-shrink-0 ml-4">
                      <span className={`text-[10px] font-black uppercase px-2 py-0.5 rounded border ${phase.cls}`}>
                        {phase.text}
                      </span>
                      <button
                        onClick={(e) => { e.stopPropagation(); onJoinRoom(room.id, room.creator_id); }}
                        className="bg-purple-600/20 hover:bg-purple-500/30 text-purple-300 hover:text-purple-200 font-black text-xs px-4 py-1.5 rounded-lg border border-purple-500/20 hover:border-purple-500/40 transition-all opacity-100 md:opacity-0 md:group-hover:opacity-100"
                      >
                        Join →
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        <p className="text-center text-slate-700 text-xs">
          Rooms without players are automatically removed after 2 hours
        </p>
      </div>
    </div>
  );
}
