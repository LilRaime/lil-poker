import { memo } from "react";

const parseCard = (cardStr: string) => {
  if (!cardStr) return null;
  const rank = cardStr.slice(0, -1);
  const suit = cardStr.slice(-1);
  const isRed = suit === "♥" || suit === "♦";
  return { rank, suit, isRed };
};

interface CardProps {
  cardStr?: string;
  faceDown?: boolean;
  className?: string;
  highlighted?: boolean;
  dimmed?: boolean;
  delayMs?: number;
  scale?: number;
}

const Card = memo(function Card({ cardStr, faceDown = false, className, highlighted = false, dimmed = false, delayMs = 0, scale }: CardProps) {
  const isSmall = className ? (
    className.includes("w-8") ||
    className.includes("w-10") ||
    className.includes("w-12")
  ) : false;
  const middleSuitSizeClass = isSmall ? "text-sm sm:text-xl" : "text-2xl sm:text-4xl";
  const sizeClasses = className || "w-12 h-18 sm:w-16 sm:h-24";

  const highlightClass = highlighted ? "ring-4 ring-yellow-400 border-yellow-400 scale-105 shadow-2xl shadow-yellow-400/50 z-20 transition-all duration-300" : "";
  const dimmedClass = dimmed ? "opacity-35 saturate-50 transition-opacity duration-300" : "";
  const hoverClass = !dimmed ? "duration-300 ease-out hover:-translate-y-1.5 hover:scale-105 hover:shadow-xl transition-all cursor-pointer" : "";

  let styleWidth: number | undefined;
  let styleHeight: number | undefined;
  let styleRankSize: number | undefined;
  let styleMiddleSuitSize: number | undefined;

  if (scale !== undefined && scale < 0.6) {
    const isPlayerCard = className?.includes("w-10");
    const baseWidth = isPlayerCard ? 40 : 48;
    const baseHeight = isPlayerCard ? 56 : 72;
    const minWidth = isPlayerCard ? 24 : 30;
    const minHeight = isPlayerCard ? 34 : 45;

    styleWidth = Math.max(minWidth, baseWidth * scale) / scale;
    styleHeight = Math.max(minHeight, baseHeight * scale) / scale;

    styleRankSize = Math.max(9, 12 * scale) / scale;
    styleMiddleSuitSize = Math.max(11, 16 * scale) / scale;
  }

  if (faceDown || !cardStr) {
    return (
      <div
        style={{
          animationDelay: `${delayMs}ms`,
          width: styleWidth ? `${styleWidth}px` : undefined,
          height: styleHeight ? `${styleHeight}px` : undefined,
        }}
        className={`${styleWidth ? "" : sizeClasses} card-back-pattern rounded-lg flex items-center justify-center relative overflow-hidden animate-deal ${dimmedClass}`}
      >
        <div className="absolute inset-1 border border-amber-500/20 rounded opacity-60 flex items-center justify-center">
          <span
            style={styleMiddleSuitSize ? { fontSize: `${styleMiddleSuitSize}px` } : undefined}
            className="text-xl sm:text-2xl text-amber-400/30 font-bold"
          >
            ♠
          </span>
        </div>
      </div>
    );
  }

  const card = parseCard(cardStr);
  if (!card) return null;

  const colorClass = card.isRed ? "text-red-600" : "text-slate-900";

  return (
    <div
      style={{
        animationDelay: `${delayMs}ms`,
        width: styleWidth ? `${styleWidth}px` : undefined,
        height: styleHeight ? `${styleHeight}px` : undefined,
      }}
      className={`${styleWidth ? "" : sizeClasses} bg-white text-slate-900 border border-slate-300 rounded-lg shadow-lg relative select-none overflow-hidden animate-deal ${highlightClass} ${dimmedClass} ${hoverClass}`}
    >
      <div className={`absolute top-0.5 left-0.5 sm:top-1 sm:left-1.5 flex flex-col items-center leading-[0.95] font-black ${colorClass}`}>
        <span style={styleRankSize ? { fontSize: `${styleRankSize}px` } : undefined} className="text-[10px] sm:text-sm">{card.rank}</span>
      </div>

      <div
        style={styleMiddleSuitSize ? { fontSize: `${styleMiddleSuitSize}px` } : undefined}
        className={`absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-center font-bold leading-none ${middleSuitSizeClass} ${colorClass}`}
      >
        {card.suit}
      </div>
      <div className={`absolute bottom-0.5 right-0.5 sm:bottom-1 sm:right-1.5 flex flex-col items-center leading-[0.95] font-black rotate-180 ${colorClass}`}>
        <span style={styleRankSize ? { fontSize: `${styleRankSize}px` } : undefined} className="text-[10px] sm:text-sm">{card.rank}</span>
      </div>
    </div>
  );
});

export default Card;
