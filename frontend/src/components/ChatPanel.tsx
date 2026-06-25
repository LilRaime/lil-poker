import { useState, useRef, useEffect, useMemo } from "react";
import { ChatMessage, HandHistoryEntry } from "../types";
import { getPlayerColor } from "../utils/playerColors";
import Card from "./Card";

function formatTime(timestamp: number): string {
  const date = new Date(timestamp);
  return `${date.getHours().toString().padStart(2, "0")}:${date.getMinutes().toString().padStart(2, "0")}`;
}

interface ChatPanelProps {
  chatMessages: ChatMessage[];
  playerId: string;
  onSendMessage: (text: string) => void;
  handHistory?: HandHistoryEntry[];
  observers?: string[];
}

export default function ChatPanel({
  chatMessages,
  playerId,
  onSendMessage,
  handHistory = [],
  observers = [],
}: ChatPanelProps) {
  const [inputText, setInputText] = useState("");
  const [activeTab, setActiveTab] = useState<"chat" | "logs" | "history">("chat");
  const [unreadChat, setUnreadChat] = useState(0);
  const [unreadLogs, setUnreadLogs] = useState(0);
  const scrollRef = useRef<HTMLDivElement | null>(null);
  const lastMsgRef = useRef<number>(0);

  const filteredMessages = useMemo(() => {
    return chatMessages.filter((msg) => {
      if (activeTab === "chat") {
        return !msg.system;
      } else {
        return msg.system;
      }
    });
  }, [chatMessages, activeTab]);

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    const text = inputText.trim();
    if (text) {
      onSendMessage(text);
      setInputText("");
      handleTabChange("chat");
    }
  };

  const handleTabChange = (tab: "chat" | "logs" | "history") => {
    setActiveTab(tab);
    if (tab === "chat") setUnreadChat(0);
    if (tab === "logs") setUnreadLogs(0);
  };

  useEffect(() => {
    if (chatMessages.length === 0) return;
    const newMessages = chatMessages.filter((msg) => msg.time > lastMsgRef.current);
    if (newMessages.length === 0) return;

    newMessages.forEach((msg) => {
      if (msg.system) {
        if (activeTab !== "logs") {
          setUnreadLogs((prev) => prev + 1);
        }
      } else {
        if (activeTab !== "chat") {
          setUnreadChat((prev) => prev + 1);
        }
      }
    });

    lastMsgRef.current = Math.max(...chatMessages.map((m) => m.time));
  }, [chatMessages, activeTab]);

  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    const isAtBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
    if (isAtBottom || activeTab === "history") {
      setTimeout(() => {
        el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
      }, 50);
    }
  }, [filteredMessages.length, activeTab]);

  return (
    <div className="flex flex-col bg-slate-900/90 border border-white/10 rounded-2xl shadow-2xl overflow-hidden h-[320px] sm:h-[385px] lg:h-[580px] w-full">
      <div className="bg-slate-950/80 border-b border-white/5 flex flex-col">
        <div className="px-4 py-3 flex items-center justify-between border-b border-white/5">
          <span className="text-xs uppercase tracking-widest text-slate-400 font-extrabold">
            Table Activity
          </span>
          <span className="text-[9px] bg-purple-500/20 text-purple-300 border border-purple-500/30 px-2 py-0.5 rounded font-black uppercase tracking-wider">
            Live
          </span>
        </div>
        {observers.length > 0 && (
          <div className="px-4 py-1.5 bg-slate-900/60 border-b border-white/5 text-[10px] text-slate-400 flex items-center gap-1.5">
            <span className="font-extrabold text-slate-500 uppercase tracking-wider flex-shrink-0">
              👁️ Observers ({observers.length}):
            </span>
            <span className="truncate text-slate-300" title={observers.join(", ")}>
              {observers.join(", ")}
            </span>
          </div>
        )}
        <div className="flex bg-slate-900/40 p-1 gap-1.5">
          <button
            type="button"
            onClick={() => handleTabChange("chat")}
            className={`flex-1 py-2 text-xs font-extrabold rounded-lg transition-all relative ${
              activeTab === "chat"
                ? "bg-purple-600/30 text-purple-300 border border-purple-500/20 shadow-inner"
                : "text-slate-400 hover:text-slate-200 hover:bg-white/5"
            }`}
          >
            💬 Chat
            {unreadChat > 0 && (
              <span className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white text-[9px] font-black flex items-center justify-center rounded-full shadow border border-slate-900">
                {unreadChat}
              </span>
            )}
          </button>
          <button
            type="button"
            onClick={() => handleTabChange("logs")}
            className={`flex-1 py-2 text-xs font-extrabold rounded-lg transition-all relative ${
              activeTab === "logs"
                ? "bg-purple-600/30 text-purple-300 border border-purple-500/20 shadow-inner"
                : "text-slate-400 hover:text-slate-200 hover:bg-white/5"
            }`}
          >
            📋 Logs
            {unreadLogs > 0 && (
              <span className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white text-[9px] font-black flex items-center justify-center rounded-full shadow border border-slate-900">
                {unreadLogs}
              </span>
            )}
          </button>
          <button
            type="button"
            onClick={() => handleTabChange("history")}
            className={`flex-1 py-2 text-xs font-extrabold rounded-lg transition-all ${
              activeTab === "history"
                ? "bg-purple-600/30 text-purple-300 border border-purple-500/20 shadow-inner"
                : "text-slate-400 hover:text-slate-200 hover:bg-white/5"
            }`}
          >
            🃏 History
          </button>
        </div>
      </div>

      {activeTab === "history" ? (
        <div ref={scrollRef} className="flex-1 overflow-y-auto p-4 space-y-3 scrollbar-thin scrollbar-thumb-white/5 scrollbar-track-transparent">
          {handHistory.length === 0 ? (
            <div className="h-full flex items-center justify-center text-center text-xs text-slate-500 italic p-4 animate-fade-in">
              No hands played yet in this session.
            </div>
          ) : (
            handHistory.map((hand, idx) => (
              <div key={`history-${hand.hand_num}-${idx}`} className="bg-slate-950/80 border border-white/5 rounded-xl p-3 flex flex-col space-y-1.5 animate-slide-up">
                <div className="flex justify-between items-center text-xxs font-black tracking-wider text-indigo-400 uppercase border-b border-white/5 pb-1">
                  <span>Hand #{hand.hand_num}</span>
                  <span className="text-[8px] bg-slate-800 text-slate-400 px-1.5 py-0.5 rounded">Showdown</span>
                </div>
                {hand.board && hand.board.length > 0 && (
                  <div className="flex items-center gap-1.5 py-1">
                    <span className="text-[9px] text-slate-500 font-extrabold uppercase mr-1">Board:</span>
                    <div className="flex gap-0.5 scale-75 origin-left">
                      {hand.board.map((c, i) => (
                        <Card key={i} cardStr={c} className="w-8 h-12" />
                      ))}
                    </div>
                  </div>
                )}
                <div className="space-y-1 border-t border-white/5 pt-1.5">
                  {hand.winners.map((winner, widx) => (
                    <div key={widx} className="text-[10px] text-slate-300 leading-tight">
                      🏆 <span className="font-extrabold text-slate-200">{winner.player_name}</span> won <span className="text-amber-400 font-black">+{winner.amount} 🪙</span>
                      {winner.hand_rank && (
                        <span className="text-purple-300 ml-1 font-medium">({winner.hand_rank})</span>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            ))
          )}
        </div>
      ) : (
        <div ref={scrollRef} className="flex-1 overflow-y-auto p-4 space-y-2.5 scrollbar-thin scrollbar-thumb-white/5 scrollbar-track-transparent">
          {filteredMessages.length === 0 ? (
            <div className="h-full flex items-center justify-center text-center text-xs text-slate-500 italic p-4 animate-fade-in">
              {activeTab === "chat"
                ? "Chat history is empty. Say hello!"
                : "No game actions logged yet."}
            </div>
          ) : (
            filteredMessages.map((msg, idx) => {
              if (msg.system) {
                return (
                  <div key={`sys-${msg.time}-${idx}`} className="text-xs text-slate-300 italic leading-snug flex items-start space-x-1.5 animate-slide-up">
                    <span className="text-[10px] text-slate-500 font-mono select-none">
                      [{formatTime(msg.time)}]
                    </span>
                    <span>📢 {msg.text}</span>
                  </div>
                );
              }

              const isMe = msg.player_id === playerId;
              return (
                <div key={`chat-${msg.time}-${idx}`} className={`text-xs leading-snug flex flex-col p-1 rounded transition-colors animate-slide-up ${isMe ? "bg-purple-950/15 border-l-2 border-purple-500/50 pl-1.5" : ""}`}>
                  <div className="flex items-baseline space-x-1.5">
                    <span className="text-[10px] text-slate-500 font-mono select-none">
                      [{formatTime(msg.time)}]
                    </span>
                    <span className={`font-black ${
                      msg.player_id
                        ? getPlayerColor(msg.player_id).text
                        : msg.player_name === "Observer"
                        ? "text-slate-400"
                        : "text-indigo-400"
                    }`}>
                      {isMe ? "You" : msg.player_name}:
                    </span>
                    <span className="text-slate-200 break-all">{msg.text}</span>
                  </div>
                </div>
              );
            })
          )}
        </div>
      )}

      <form onSubmit={handleSend} className="bg-slate-950/80 p-2 border-t border-white/5 flex gap-2">
        <input
          type="text"
          value={inputText}
          onChange={(e) => setInputText(e.target.value)}
          placeholder={activeTab === "logs" || activeTab === "history" ? "Send a chat message (switches to Chat)..." : "Send a message..."}
          maxLength={100}
          className="flex-1 bg-slate-900 border border-white/5 rounded-xl px-3.5 py-2 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-purple-500/50 transition-colors"
        />
        <button
          type="submit"
          disabled={!inputText.trim()}
          className="bg-purple-600 hover:bg-purple-500 disabled:bg-purple-950/45 disabled:text-purple-300/30 font-bold px-4 py-2 rounded-xl text-sm text-white transition-colors"
        >
          Send
        </button>
      </form>
    </div>
  );
}
