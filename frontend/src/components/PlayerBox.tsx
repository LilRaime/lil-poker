import { useEffect, useState, memo } from "react";
import { PlayerStatus } from "../types";
import Card from "./Card";
import { getPlayerColor } from "../utils/playerColors";

interface PlayerBoxProps {
  player: PlayerStatus;
  isPlayerActive: boolean;
  isDealer: boolean;
  isSmallBlind: boolean;
  isBigBlind: boolean;
  phase: string;
  actionDeadline?: number;
  winningCards?: string[];
  dealDelayStart?: number;
  activePlayerCount?: number;
  positionIndex: number;
  scale?: number;
}

const getLayoutConfig = (positionIndex: number) => {
  if (positionIndex === 0 || positionIndex === 1 || positionIndex === 7) {
    let betClass = "absolute bottom-[calc(100%+32px)] right-[calc(50%+56px)] z-20";
    let handRankClass = "absolute top-[calc(100%+6px)] left-1/2 -translate-x-1/2 z-30 whitespace-nowrap text-center";
    if (positionIndex === 1) {
      betClass = "absolute bottom-[calc(100%+32px)] left-[calc(50%+56px)] z-20";
      handRankClass = "absolute top-[calc(100%+6px)] left-0 z-30 whitespace-nowrap text-center";
    } else if (positionIndex === 7) {
      betClass = "absolute bottom-[calc(100%+32px)] right-[calc(50%+56px)] z-20";
      handRankClass = "absolute top-[calc(100%+6px)] right-0 z-30 whitespace-nowrap text-center";
    }
    return {
      cards: "absolute bottom-[calc(100%+12px)] left-1/2 -translate-x-1/2 flex space-x-1 z-20",
      handRank: handRankClass,
      bet: betClass,
      tooltip: "absolute bottom-full mb-3 left-1/2 -translate-x-1/2 z-30",
      reaction: "absolute -bottom-16 left-1/2 -translate-x-1/2 bg-slate-900/90 border border-purple-500/50 rounded-full px-3 py-1 text-lg shadow-xl animate-bounce z-30",
    };
  }

  if (positionIndex === 3 || positionIndex === 4 || positionIndex === 5) {
    let betClass = "absolute top-[calc(100%+32px)] left-[calc(50%+56px)] z-20";
    let handRankClass = "absolute bottom-[calc(100%+6px)] left-1/2 -translate-x-1/2 z-30 whitespace-nowrap text-center";
    if (positionIndex === 3) {
      betClass = "absolute top-[calc(100%+32px)] left-[calc(50%+56px)] z-20";
      handRankClass = "absolute bottom-[calc(100%+6px)] left-0 z-30 whitespace-nowrap text-center";
    } else if (positionIndex === 5) {
      betClass = "absolute top-[calc(100%+32px)] right-[calc(50%+56px)] z-20";
      handRankClass = "absolute bottom-[calc(100%+6px)] right-0 z-30 whitespace-nowrap text-center";
    }
    return {
      cards: "absolute top-[calc(100%+12px)] left-1/2 -translate-x-1/2 flex space-x-1 z-20",
      handRank: handRankClass,
      bet: betClass,
      tooltip: "absolute top-full mt-3 left-1/2 -translate-x-1/2 z-30",
      reaction: "absolute -top-12 left-1/2 -translate-x-1/2 bg-slate-900/90 border border-purple-500/50 rounded-full px-3 py-1 text-lg shadow-xl animate-bounce z-30",
    };
  }

  if (positionIndex === 2) {
    return {
      cards: "absolute left-[calc(100%+8px)] top-1/2 -translate-y-1/2 flex space-x-1 z-20",
      handRank: "absolute bottom-[calc(100%+6px)] left-0 z-30 whitespace-nowrap text-center",
      bet: "absolute left-[calc(100%+116px)] top-1/2 -translate-y-1/2 z-20",
      tooltip: "absolute bottom-full mb-3 left-1/2 -translate-x-1/2 z-30",
      reaction: "absolute -top-12 left-1/2 -translate-x-1/2 bg-slate-900/90 border border-purple-500/50 rounded-full px-3 py-1 text-lg shadow-xl animate-bounce z-30",
    };
  }

  return {
    cards: "absolute right-[calc(100%+8px)] top-1/2 -translate-y-1/2 flex space-x-1 z-20",
    handRank: "absolute bottom-[calc(100%+6px)] right-0 z-30 whitespace-nowrap text-center",
    bet: "absolute right-[calc(100%+116px)] top-1/2 -translate-y-1/2 z-20",
    tooltip: "absolute bottom-full mb-3 left-1/2 -translate-x-1/2 z-30",
    reaction: "absolute -top-12 left-1/2 -translate-x-1/2 bg-slate-900/90 border border-purple-500/50 rounded-full px-3 py-1 text-lg shadow-xl animate-bounce z-30",
  };
};

const PlayerBoxComponent = function PlayerBox({
  player,
  isPlayerActive,
  isDealer,
  isSmallBlind,
  isBigBlind,
  phase,
  actionDeadline,
  winningCards,
  dealDelayStart = 0,
  activePlayerCount = 1,
  positionIndex,
  scale = 1,
}: PlayerBoxProps) {
  const [timeLeft, setTimeLeft] = useState<number>(0);
  const [totalTime, setTotalTime] = useState<number>(20);

  useEffect(() => {
    if (!isPlayerActive || !actionDeadline) {
      setTimeLeft(0);
      return;
    }

    const remainingMs = actionDeadline - Date.now();
    const remainingSeconds = Math.max(1, Math.round(remainingMs / 1000));
    setTotalTime(remainingSeconds);

    const updateTimer = () => {
      const remaining = Math.max(0, Math.round((actionDeadline - Date.now()) / 1000));
      setTimeLeft(remaining);
    };

    updateTimer();
    const interval = setInterval(updateTimer, 500);

    return () => clearInterval(interval);
  }, [isPlayerActive, actionDeadline]);

  const color = getPlayerColor(player.id);
  const cfg = getLayoutConfig(positionIndex);

  let nameStyle: React.CSSProperties = {};
  let chipsStyle: React.CSSProperties = {};
  let badgeStyle: React.CSSProperties = {};
  let dealerBtnStyle: React.CSSProperties = {};
  let smallBlindStyle: React.CSSProperties = {};
  let bigBlindStyle: React.CSSProperties = {};
  let betLabelStyle: React.CSSProperties = {};
  let handRankLabelStyle: React.CSSProperties = {};

  if (scale !== 1) {
    nameStyle = {
      fontSize: `${Math.max(9, 14 * scale) / scale}px`,
    };

    chipsStyle = {
      fontSize: `${Math.max(8.5, 12 * scale) / scale}px`,
    };

    badgeStyle = {
      fontSize: `${Math.max(7.5, 9.5 * scale) / scale}px`,
      paddingLeft: `${Math.max(3, 6 * scale) / scale}px`,
      paddingRight: `${Math.max(3, 6 * scale) / scale}px`,
      paddingTop: `${Math.max(1, 2 * scale) / scale}px`,
      paddingBottom: `${Math.max(1, 2 * scale) / scale}px`,
      marginTop: `${Math.max(2, 6 * scale) / scale}px`,
    };

    const dSize = Math.max(16, 24 * scale) / scale;
    const dOffset = -8 / scale;
    dealerBtnStyle = {
      width: `${dSize}px`,
      height: `${dSize}px`,
      fontSize: `${Math.max(8, 10 * scale) / scale}px`,
      top: `${dOffset}px`,
      right: `${dOffset}px`,
    };
    smallBlindStyle = {
      width: `${dSize}px`,
      height: `${dSize}px`,
      fontSize: `${Math.max(8, 10 * scale) / scale}px`,
      top: `${dOffset}px`,
      left: `${dOffset}px`,
    };
    bigBlindStyle = {
      width: `${dSize}px`,
      height: `${dSize}px`,
      fontSize: `${Math.max(8, 10 * scale) / scale}px`,
      top: `${dOffset}px`,
      left: `${dOffset}px`,
    };

    betLabelStyle = {
      fontSize: `${Math.max(9.5, 11 * scale) / scale}px`,
      paddingLeft: `${Math.max(6, 10 * scale) / scale}px`,
      paddingRight: `${Math.max(6, 10 * scale) / scale}px`,
      paddingTop: `${Math.max(2, 4 * scale) / scale}px`,
      paddingBottom: `${Math.max(2, 4 * scale) / scale}px`,
    };

    handRankLabelStyle = {
      fontSize: `${Math.max(8.5, 9.5 * scale) / scale}px`,
      paddingLeft: `${Math.max(6, 10 * scale) / scale}px`,
      paddingRight: `${Math.max(6, 10 * scale) / scale}px`,
      paddingTop: `${Math.max(2, 4 * scale) / scale}px`,
      paddingBottom: `${Math.max(2, 4 * scale) / scale}px`,
      borderRadius: `${Math.max(4, 8 * scale) / scale}px`,
    };
  }

  return (
    <div className="flex flex-col items-center group relative">
      <div className={`${cfg.tooltip} group-hover:opacity-100 opacity-0 pointer-events-none bg-slate-900/95 border border-purple-500/30 rounded-xl p-4 shadow-2xl z-30 text-xs font-bold text-slate-300 w-52 transition-all duration-200 flex flex-col space-y-2`}>
        <div className="text-slate-100 border-b border-white/5 pb-1.5 font-black uppercase text-center tracking-wider text-[10px]">
          Session Stats
        </div>
        <div className="flex justify-between">
          <span className="text-slate-500">Hands Played:</span>
          <span>{player.hands_played}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-slate-500">VPIP:</span>
          <span>
            {player.hands_played > 0
              ? `${Math.round((player.hands_vpip / player.hands_played) * 100)}%`
              : "0%"}
          </span>
        </div>
        <div className="flex justify-between">
          <span className="text-slate-500">Max Pot Won:</span>
          <span className="text-amber-400 font-mono">{player.biggest_pot_won}</span>
        </div>
      </div>

      {player.reaction && (
        <div className={cfg.reaction}>
          {player.reaction}
        </div>
      )}

      <div
        style={scale !== 1 ? {
          width: `${(Math.max(80, 128 * scale) / scale)}px`,
          height: `${(Math.max(48, 76 * scale) / scale)}px`,
        } : undefined}
        className={`w-32 sm:w-36 rounded-2xl p-2.5 flex flex-col items-center justify-center transition-all shadow-xl relative ${isPlayerActive
          ? `bg-purple-950/80 border-2 ${color.border} active-turn-glow text-white`
          : player.sitting_out
            ? "bg-slate-950/60 border border-white/5 opacity-40 text-slate-500"
            : player.folded
              ? "bg-slate-900/40 border border-white/5 opacity-60 text-slate-500"
              : `bg-slate-900/90 border ${color.border} border-opacity-30 text-slate-200`
          }`}
      >
        {isDealer && (
          <div style={dealerBtnStyle} className="absolute -top-2 -right-2 w-6 h-6 bg-yellow-500 text-slate-950 text-xxs font-black flex items-center justify-center rounded-full border border-slate-900 shadow-md animate-bet-pop" title="Dealer">
            D
          </div>
        )}
        {isSmallBlind && (
          <div style={smallBlindStyle} className="absolute -top-2 -left-2 w-6 h-6 bg-blue-500 text-white text-xxs font-black flex items-center justify-center rounded-full border border-slate-900 shadow-md animate-bet-pop" title="Small Blind">
            SB
          </div>
        )}
        {isBigBlind && (
          <div style={bigBlindStyle} className="absolute -top-2 -left-2 w-6 h-6 bg-red-500 text-white text-xxs font-black flex items-center justify-center rounded-full border border-slate-900 shadow-md animate-bet-pop" title="Big Blind">
            BB
          </div>
        )}

        <div className="flex items-center gap-1.5 max-w-full">
          <span className={`w-2 h-2 rounded-full flex-shrink-0 ${color.dot}`} />
          <span style={nameStyle} className="text-sm sm:text-base font-extrabold truncate leading-none">
            {player.name}
          </span>
        </div>
        <span style={chipsStyle} className="text-xs sm:text-sm font-bold mt-1 text-slate-300 flex items-center whitespace-nowrap">
          {player.chips} 🪙
        </span>

        {player.sitting_out ? (
          <span style={badgeStyle} className="text-[9px] sm:text-[10px] uppercase tracking-widest bg-amber-600/30 text-amber-300 border border-amber-500/20 px-1.5 py-0.5 rounded mt-1.5 font-bold">
            SIT OUT
          </span>
        ) : player.all_in ? (
          <span style={badgeStyle} className="text-[9px] sm:text-[10px] uppercase tracking-widest bg-red-600/30 text-red-300 border border-red-500/20 px-1.5 py-0.5 rounded mt-1.5 font-bold">
            ALL-IN
          </span>
        ) : player.folded ? (
          <span style={badgeStyle} className={`text-[9px] sm:text-[10px] uppercase tracking-widest px-1.5 py-0.5 rounded mt-1.5 font-bold ${
            player.exposed_cards
              ? "bg-purple-600/30 text-purple-300 border border-purple-500/20"
              : "bg-slate-800 text-slate-400"
          }`}>
            {player.exposed_cards ? "SHOWED" : "FOLD"}
          </span>
        ) : player.acted ? (
          <span style={badgeStyle} className="text-[9px] sm:text-[10px] uppercase tracking-widest bg-emerald-600/30 text-emerald-300 border border-emerald-500/20 px-1.5 py-0.5 rounded mt-1.5 font-bold">
            Acted
          </span>
        ) : null}

        {isPlayerActive && timeLeft > 0 && (
          <div className="w-full bg-slate-990/60 h-1 rounded-full mt-2 overflow-hidden border border-white/5 shadow-inner">
            <div
              className={`h-full transition-all duration-500 ease-linear ${timeLeft <= 5 ? "bg-red-500 animate-pulse" : timeLeft <= 10 ? "bg-amber-500" : "bg-purple-500"
                }`}
              style={{ width: `${(timeLeft / totalTime) * 100}%` }}
            />
          </div>
        )}
      </div>

      <div className={cfg.cards}>
        {player.hole && player.hole.length > 0 ? (
          player.hole.map((card, cardIdx) => {
            const isHighlighted = winningCards?.includes(card);
            const isDimmed = winningCards && winningCards.length > 0 && !winningCards.includes(card);
            const delayMs = dealDelayStart + cardIdx * activePlayerCount * 150;
            return (
              <Card
                key={cardIdx}
                cardStr={card}
                className="w-10 h-14 sm:w-12 sm:h-18"
                highlighted={isHighlighted}
                dimmed={isDimmed}
                delayMs={delayMs}
                scale={scale}
              />
            );
          })
        ) : !player.folded && phase !== "Waiting" ? (
          <>
            <Card faceDown className="w-10 h-14 sm:w-12 sm:h-18" dimmed={winningCards && winningCards.length > 0} delayMs={dealDelayStart} scale={scale} />
            <Card faceDown className="w-10 h-14 sm:w-12 sm:h-18" dimmed={winningCards && winningCards.length > 0} delayMs={dealDelayStart + activePlayerCount * 150} scale={scale} />
          </>
        ) : null}
      </div>

      {player.current_hand && (
        <div style={handRankLabelStyle} className={`${cfg.handRank} bg-gradient-to-r from-purple-800/90 to-indigo-800/90 border border-purple-400/40 text-purple-100 px-2.5 py-1 rounded-lg text-[10px] sm:text-xs font-black shadow-md uppercase tracking-wider animate-fade-in text-center max-w-none leading-tight flex items-center justify-center gap-1`}>
          <span>⭐️</span> <span>{player.current_hand}</span>
        </div>
      )}

      {player.bet > 0 && (
        <div
          style={scale !== 1 ? betLabelStyle : undefined}
          className={`${cfg.bet} bg-amber-500/20 border border-amber-500/30 text-amber-300 font-black text-xs px-2.5 py-1 rounded-full flex items-center gap-1 shadow-2xl backdrop-blur-sm animate-bet-pop`}
        >
          <span>💰</span>
          <span>{player.bet}</span>
        </div>
      )}
    </div>
  );
};

const PlayerBox = memo(PlayerBoxComponent);

export default PlayerBox;
