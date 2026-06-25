import { useEffect } from "react";

interface ConfirmModalProps {
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
}

export default function ConfirmModal({
  title,
  message,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  onConfirm,
  onCancel,
}: ConfirmModalProps) {
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onCancel();
      if (e.key === "Enter") onConfirm();
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [onCancel, onConfirm]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm animate-fade-in"
      onClick={onCancel}
    >
      <div
        className="glass-panel-heavy border border-white/10 rounded-2xl shadow-2xl p-6 max-w-sm w-full mx-4 flex flex-col gap-4"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-3">
          <span className="text-2xl">⚠️</span>
          <h2 className="text-base font-black text-slate-100">{title}</h2>
        </div>
        <p className="text-sm text-slate-400 leading-relaxed">{message}</p>
        <div className="flex gap-2 mt-1">
          <button
            onClick={onCancel}
            autoFocus
            className="flex-1 py-2.5 bg-slate-800 hover:bg-slate-700 border border-white/5 font-bold rounded-xl text-sm transition-colors"
          >
            {cancelLabel}
          </button>
          <button
            onClick={onConfirm}
            className="flex-1 py-2.5 bg-red-700 hover:bg-red-600 font-bold rounded-xl text-sm transition-colors text-white shadow"
          >
            {confirmLabel}
          </button>
        </div>
        <p className="text-center text-xxxxs text-slate-600">
          Enter to confirm · Esc to cancel
        </p>
      </div>
    </div>
  );
}
