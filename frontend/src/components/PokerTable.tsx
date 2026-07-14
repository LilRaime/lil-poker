import { memo, useMemo, useRef, useState, useEffect } from "react";
import { GameStateResponse, PlayerStatus } from "../types";
import Card from "./Card";
import PlayerBox from "./PlayerBox";
import { useFlyingChips } from "../hooks/useFlyingChips";

const EmptyCardSlot = memo(function EmptyCardSlot({ scale = 1 }: { scale?: number }) {
  const width = Math.max(32, 48 * scale) / scale;
  const height = Math.max(48, 72 * scale) / scale;
  const fontSize = Math.max(8, 12 * scale) / scale;
  return (
    <div
      style={{ width: `${width}px`, height: `${height}px` }}
      className="border-2 border-dashed border-white/10 rounded-lg flex items-center justify-center"
    >
      <span style={{ fontSize: `${fontSize}px` }} className="text-white/5 font-semibold">Card</span>
    </div>
  );
});

const PLAYER_POSITIONS = [
  "-translate-x-1/2",
  "",
  "-translate-y-1/2",
  "",
  "-translate-x-1/2",
  "",
  "-translate-y-1/2",
  "",
];

const getSeatStyle = (index: number, scale: number): React.CSSProperties => {
  const isMobile = scale !== 1;
  const factor = isMobile ? 1 - scale : 0;

  const edgeOffsetVertical = - (72 + 50 * factor) / scale;
  const sideSeatOffset = - (92 + 35 * factor) / scale;

  const cornerY = - (52 + 40 * factor) / scale;
  const cornerX = (15 - 45 * factor) / scale;

  const clampedSideEdge = isMobile ? Math.max(-48, sideSeatOffset) : sideSeatOffset;
  const clampedCornerX = isMobile ? Math.max(-36, cornerX) : cornerX;

  switch (index) {
    case 0:
      return { bottom: `${edgeOffsetVertical}px`, left: "50%" };
    case 1:
      return { bottom: `${cornerY}px`, left: `${clampedCornerX}px` };
    case 2:
      return { top: "50%", left: `${clampedSideEdge}px` };
    case 3:
      return { top: `${cornerY}px`, left: `${clampedCornerX}px` };
    case 4:
      return { top: `${edgeOffsetVertical}px`, left: "50%" };
    case 5:
      return { top: `${cornerY}px`, right: `${clampedCornerX}px` };
    case 6:
      return { top: "50%", right: `${clampedSideEdge}px` };
    case 7:
      return { bottom: `${cornerY}px`, right: `${clampedCornerX}px` };
    default:
      return {};
  }
};

interface PokerTableProps {
  gameState: GameStateResponse;
  playerId: string;
  onJoin: (seat: number) => void;
  winningCards: string[];
  isShowdown: boolean;
  isCreator: boolean;
  onKick: (uuid: string) => void;
}

export default function PokerTable({
  gameState,
  playerId,
  onJoin,
  winningCards,
  isShowdown,
  isCreator,
  onKick,
}: PokerTableProps) {
  const { flyingChips, removeChip } = useFlyingChips(gameState, playerId);

  const [scale, setScale] = useState(1);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleResize = () => {
      if (!containerRef.current) return;
      const parentWidth = containerRef.current.clientWidth;
      const targetWidth = 840;
      if (parentWidth > 0) {
        if (parentWidth < targetWidth) {
          setScale(parentWidth / targetWidth);
        } else {
          setScale(1);
        }
      }
    };

    handleResize();
    if (typeof ResizeObserver !== "undefined" && containerRef.current) {
      const observer = new ResizeObserver(handleResize);
      observer.observe(containerRef.current);
      return () => observer.disconnect();
    }

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  const wrapperHeight = 520;

  const hero = useMemo(() => gameState.players.find((p) => p.id === playerId), [gameState.players, playerId]);
  const heroSeat = hero ? hero.seat : 0;

  const rotateSeat = (seatIdx: number) => {
    if (!hero) return seatIdx;
    return (seatIdx - heroSeat + 8) % 8;
  };

  const playersBySeat = useMemo(() => {
    const map = new Map<number, PlayerStatus>();
    gameState.players.forEach((p) => map.set(p.seat, p));
    return map;
  }, [gameState.players]);

  const activePlayers = useMemo(() => {
    return gameState.players
      .filter((p) => !p.sitting_out && !p.folded && p.chips >= 0)
      .sort((a, b) => a.seat - b.seat);
  }, [gameState.players]);

  const extraHeight = scale === 1 ? 20 : Math.round((1 - scale) * 270);

  return (
    <div
      ref={containerRef}
      className="w-full max-w-3xl flex items-center justify-center my-1.5 sm:my-2 relative"
      style={{ height: `${(wrapperHeight + extraHeight) * scale}px` }}
    >
      <div
        className="select-none"
        style={{
          transform: `scale(${scale})`,
          transformOrigin: "top left",
          position: "absolute",
          top: 0,
          left: 0,
          width: "840px",
          height: "520px",
        }}
      >
        <div
          className="w-[768px] h-[384px] poker-felt rounded-full absolute p-4 flex items-center justify-center"
          style={{
            top: "68px",
            left: "36px",
          }}
        >
          <div className="absolute inset-2 border-2 border-white/5 rounded-full pointer-events-none opacity-20" />

          <div className="flex flex-col items-center glass-panel px-4 py-3.5 rounded-xl shadow-2xl border border-white/5 z-0">
            <div className="flex items-center space-x-2 bg-slate-900/80 border border-amber-500/20 px-4 py-2 rounded-full mb-2 shadow-inner">
              <span className="text-yellow-400 text-base">🪙</span>
              <span className="text-[10px] uppercase tracking-widest text-slate-400">Pot:</span>
              <span key={gameState.pot} className="font-black text-amber-100 text-base animate-pot-pop">
                {gameState.pot}
              </span>
            </div>

            {gameState.sub_pots && gameState.sub_pots.length > 1 && (
              <div className="flex flex-wrap justify-center gap-1.5 mb-2 max-w-[280px]">
                {gameState.sub_pots.map((sp, idx) => (
                  <div
                    key={idx}
                    className="text-[10px] bg-slate-800/80 border border-white/5 px-2 py-0.5 rounded-md text-slate-300 flex items-center space-x-1 cursor-help group relative shadow-sm"
                  >
                    <span className="font-semibold text-yellow-400/90">
                      {idx === 0 ? "Main" : `Side ${idx}`}:
                    </span>
                    <span>{sp.amount}</span>
                    <div className="invisible group-hover:visible absolute bottom-full left-1/2 transform -translate-x-1/2 mb-1.5 px-2 py-1 bg-slate-950/95 text-slate-200 text-[9px] rounded shadow-xl whitespace-nowrap z-50 border border-white/10 pointer-events-none">
                      {sp.contributors.join(", ")}
                    </div>
                  </div>
                ))}
              </div>
            )}

            <div className="flex space-x-2 min-h-[96px] items-center justify-center">
              {gameState.board && gameState.board.length > 0 ? (
                <>
                  {gameState.board.map((card, i) => {
                    const isHighlighted = isShowdown && winningCards.includes(card);
                    const isDimmed = isShowdown && !winningCards.includes(card);
                    return (
                      <Card
                        key={i}
                        cardStr={card}
                        highlighted={isHighlighted}
                        dimmed={isDimmed}
                        delayMs={i * 150}
                        scale={scale}
                      />
                    );
                  })}
                  {Array.from({ length: 5 - gameState.board.length }).map((_, i) => (
                    <EmptyCardSlot key={i} scale={scale} />
                  ))}
                </>
              ) : (
                Array.from({ length: 5 }).map((_, i) => (
                  <EmptyCardSlot key={i} scale={scale} />
                ))
              )}
            </div>
          </div>

          {Array.from({ length: 8 }).map((_, seatIdx) => {
            const player = playersBySeat.get(seatIdx);
            const positionIndex = rotateSeat(seatIdx);
            const positionClass = PLAYER_POSITIONS[positionIndex];

            const activeIdx = player ? activePlayers.findIndex((p) => p.id === player.id) : -1;
            const dealDelayStart = activeIdx !== -1 ? activeIdx * 150 : 0;

            const isDealer = player ? gameState.players[gameState.dealer_idx]?.id === player.id : false;
            const isSmallBlind = player ? !!player.is_small_blind : false;
            const isBigBlind = player ? !!player.is_big_blind : false;

            const seatStyle = getSeatStyle(positionIndex, scale);

            return (
              <div
                key={`seat-${seatIdx}`}
                style={seatStyle}
                className={`absolute flex flex-col items-center z-10 ${positionClass}`}
              >
                {player ? (
                  <PlayerBox
                    player={player}
                    isPlayerActive={
                      (gameState.active_player_id
                        ? gameState.active_player_id === player.id
                        : gameState.players[gameState.active_idx]?.id === player.id) &&
                      gameState.phase !== "Waiting"
                    }
                    isDealer={isDealer}
                    isSmallBlind={gameState.phase !== "Waiting" && isSmallBlind}
                    isBigBlind={gameState.phase !== "Waiting" && isBigBlind}
                    phase={gameState.phase}
                    actionDeadline={gameState.action_deadline}
                    winningCards={winningCards}
                    dealDelayStart={dealDelayStart}
                    activePlayerCount={activePlayers.length}
                    positionIndex={positionIndex}
                    scale={scale}
                    isCreator={isCreator}
                    onKick={onKick}
                    currentUserId={playerId}
                  />
                ) : (
                  <div
                    style={scale !== 1 ? {
                      width: `${(Math.max(66, 112 * scale) / scale)}px`,
                      height: `${(Math.max(75, 125 * scale) / scale)}px`,
                    } : undefined}
                    className="w-28 sm:w-32 h-[125px] flex items-center justify-center"
                  >
                    <div
                      style={scale !== 1 ? {
                        width: `${(Math.max(62, 96 * scale) / scale)}px`,
                        padding: `${(Math.max(4, 10 * scale) / scale)}px`,
                        borderRadius: `${(Math.max(6, 16 * scale) / scale)}px`,
                      } : undefined}
                      className="w-24 sm:w-28 rounded-2xl border border-dashed border-white/10 bg-slate-950/20 p-2.5 flex flex-col items-center justify-center text-center backdrop-blur-sm transition-all duration-300"
                    >
                      <span
                        style={scale !== 1 ? {
                          fontSize: `${Math.max(8, 10 * scale) / scale}px`,
                          marginBottom: `${Math.max(3, 6 * scale) / scale}px`,
                        } : undefined}
                        className="text-[10px] text-slate-500 font-extrabold uppercase tracking-wider mb-1.5"
                      >
                        Seat {seatIdx + 1}
                      </span>
                      {!hero ? (
                        <button
                          onClick={() => onJoin(seatIdx)}
                          style={scale !== 1 ? {
                            fontSize: `${Math.max(7.5, 9.5 * scale) / scale}px`,
                            paddingLeft: `${Math.max(4, 8 * scale) / scale}px`,
                            paddingRight: `${Math.max(4, 8 * scale) / scale}px`,
                            paddingTop: `${Math.max(3, 6 * scale) / scale}px`,
                            paddingBottom: `${Math.max(3, 6 * scale) / scale}px`,
                            borderRadius: `${Math.max(4, 8 * scale) / scale}px`,
                          } : undefined}
                          className="text-xxxxs sm:text-xxs px-2 py-1.5 rounded-lg bg-purple-600/20 hover:bg-purple-500/40 text-purple-300 border border-purple-500/20 hover:scale-105 active:scale-95 transition-all font-black uppercase tracking-wider shadow-md"
                        >
                          + Sit Here
                        </button>
                      ) : (
                        <span
                          style={scale !== 1 ? {
                            fontSize: `${Math.max(8, 10 * scale) / scale}px`,
                          } : undefined}
                          className="text-xxs text-slate-600 italic font-bold uppercase tracking-wider"
                        >
                          Empty
                        </span>
                      )}
                    </div>
                  </div>
                )}
              </div>
            );
          })}

          {flyingChips.map((chip) => (
            <div
              key={chip.id}
              className="absolute w-4 h-4 bg-amber-500 rounded-full border border-amber-600 shadow-md flex items-center justify-center font-bold text-[8px] text-amber-950 select-none pointer-events-none z-50 animate-chip-fly"
              style={{
                "--start-x": `${chip.startX}%`,
                "--start-y": `${chip.startY}%`,
                "--end-x": `${chip.endX}%`,
                "--end-y": `${chip.endY}%`,
                animationDelay: `${chip.delay}ms`,
              } as React.CSSProperties}
              onAnimationEnd={() => {
                removeChip(chip.id);
              }}
            >
              🪙
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
