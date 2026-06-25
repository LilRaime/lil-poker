import { PlayerStatus } from "../types";

interface SitOutBannerProps {
  hero: PlayerStatus;
  onSitIn: () => void;
  onRebuyAndSitIn: () => void;
}

export default function SitOutBanner({ hero, onSitIn, onRebuyAndSitIn }: SitOutBannerProps) {
  const isOutOfChips = hero.chips === 0;

  return (
    <div className="mb-6 w-full max-w-xl text-center glass-panel px-6 py-5 rounded-2xl flex flex-col sm:flex-row items-center justify-between gap-4 border border-amber-500/20 shadow-xl bg-amber-950/10 animate-fade-in">
      <div className="text-left">
        <h3 className="font-bold text-amber-200">
          {isOutOfChips ? "You are out of chips!" : "You are Sitting Out"}
        </h3>
        <p className="text-xs text-slate-400 mt-0.5">
          {isOutOfChips
            ? "Claim a Rebuy to get back in the game and sit back in."
            : "You won't be dealt in until you sit back in."}
        </p>
      </div>

      {isOutOfChips ? (
        <button
          onClick={onRebuyAndSitIn}
          className="w-full sm:w-auto px-5 py-2.5 bg-emerald-600 hover:bg-emerald-500 font-extrabold rounded-xl text-sm shadow transition-colors text-white"
        >
          Claim Rebuy &amp; Sit In
        </button>
      ) : (
        <button
          onClick={onSitIn}
          className="w-full sm:w-auto px-5 py-2.5 bg-amber-600 hover:bg-amber-500 font-extrabold rounded-xl text-sm shadow transition-colors text-slate-950"
        >
          I&apos;m Back / Sit In
        </button>
      )}
    </div>
  );
}
