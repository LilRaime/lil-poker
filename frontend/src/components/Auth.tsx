import React, { useState } from "react";

interface AuthProps {
  onAuthSuccess: (uuid: string, username: string, chips: number, isGuest?: boolean, guestPassword?: string) => void;
}

export default function Auth({ onAuthSuccess }: AuthProps) {
  const [mode, setMode] = useState<"login" | "register" | "guest">("login");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const trimmedUsername = username.trim();
    if (trimmedUsername.length < 3) {
      setError("Username must be at least 3 characters long");
      return;
    }

    if (mode !== "guest") {
      if (password.length < 4) {
        setError("Password must be at least 4 characters long");
        return;
      }

      if (mode === "register" && password !== confirmPassword) {
        setError("Passwords do not match");
        return;
      }
    }

    setLoading(true);

    let endpoint = "/api/auth/login";
    const bodyData: any = { username: trimmedUsername };

    if (mode === "register") {
      endpoint = "/api/auth/register";
      bodyData.password = password;
    } else if (mode === "guest") {
      endpoint = "/api/auth/guest";
    } else {
      bodyData.password = password;
    }

    try {
      const res = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(bodyData),
      });

      const data = await res.json();

      if (!res.ok) {
        throw new Error(data.error || "Authentication failed");
      }

      onAuthSuccess(data.uuid, data.username, data.chips || 1000, mode === "guest", data.password);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-4">
      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        <div className="absolute top-1/4 left-1/2 -translate-x-1/2 w-[600px] h-[600px] rounded-full bg-purple-600/5 blur-3xl" />
      </div>

      <div className="w-full max-w-md glass-panel p-8 rounded-2xl shadow-2xl relative overflow-hidden border border-white/10 animate-fade-in">
        <div className="absolute -top-12 -right-12 w-36 h-36 bg-purple-600/10 blur-3xl rounded-full" />
        <div className="absolute -bottom-12 -left-12 w-36 h-36 bg-indigo-600/10 blur-3xl rounded-full" />

        <div className="text-center mb-8">
          <div className="text-5xl mb-4 animate-bounce">♠️ ♥️ ♦️ ♣️</div>
          <h2 className="text-2xl font-black tracking-wide text-slate-100">
            {mode === "login" ? "Welcome Back" : mode === "register" ? "Create Account" : "Play Instantly"}
          </h2>
          <p className="text-slate-400 text-sm mt-1">
            {mode === "login"
              ? "Sign in to play real-time poker"
              : mode === "register"
                ? "Sign up to start with 1,000 chips"
                : "Enter a nickname to start with 1,000 chips"}
          </p>
        </div>

        <div className="flex bg-slate-950/60 p-1 rounded-xl mb-6 border border-white/5">
          <button
            type="button"
            onClick={() => {
              setMode("login");
              setError(null);
              setPassword("");
              setConfirmPassword("");
            }}
            className={`flex-1 py-2 text-xs font-bold rounded-lg transition-all ${mode === "login"
                ? "bg-gradient-to-r from-purple-600 to-indigo-600 text-white shadow-md"
                : "text-slate-400 hover:text-slate-200"
              }`}
          >
            Login
          </button>
          <button
            type="button"
            onClick={() => {
              setMode("register");
              setError(null);
              setPassword("");
              setConfirmPassword("");
            }}
            className={`flex-1 py-2 text-xs font-bold rounded-lg transition-all ${mode === "register"
                ? "bg-gradient-to-r from-purple-600 to-indigo-600 text-white shadow-md"
                : "text-slate-400 hover:text-slate-200"
              }`}
          >
            Register
          </button>
          <button
            type="button"
            onClick={() => {
              setMode("guest");
              setError(null);
              setPassword("");
              setConfirmPassword("");
            }}
            className={`flex-1 py-2 text-xs font-bold rounded-lg transition-all ${mode === "guest"
                ? "bg-gradient-to-r from-purple-600 to-indigo-600 text-white shadow-md"
                : "text-slate-400 hover:text-slate-200"
              }`}
          >
            Guest
          </button>
        </div>

        {error && (
          <div className="mb-4 bg-red-950/40 border border-red-500/20 text-red-200 text-xs px-4 py-3 rounded-xl flex items-center space-x-2 font-semibold">
            <span>⚠️</span>
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xxs uppercase tracking-widest text-slate-400 mb-1.5 font-bold">
              Username
            </label>
            <input
              type="text"
              required
              maxLength={12}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Username"
              className="w-full bg-slate-900 border border-white/10 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 rounded-xl px-4 py-3 text-sm font-semibold transition-all focus:outline-none text-white"
            />
          </div>

          {mode !== "guest" && (
            <div>
              <label className="block text-xxs uppercase tracking-widest text-slate-400 mb-1.5 font-bold">
                Password
              </label>
              <input
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full bg-slate-900 border border-white/10 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 rounded-xl px-4 py-3 text-sm font-semibold transition-all focus:outline-none text-white"
              />
            </div>
          )}

          {mode === "register" && (
            <div className="animate-slide-down">
              <label className="block text-xxs uppercase tracking-widest text-slate-400 mb-1.5 font-bold">
                Confirm Password
              </label>
              <input
                type="password"
                required
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full bg-slate-900 border border-white/10 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 rounded-xl px-4 py-3 text-sm font-semibold transition-all focus:outline-none text-white"
              />
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full mt-2 bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 py-3.5 rounded-xl font-bold tracking-wide shadow-lg hover:shadow-indigo-500/20 active:translate-y-px transition-all text-white flex justify-center items-center"
          >
            {loading ? (
              <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
            ) : mode === "login" ? (
              "Login to Play"
            ) : mode === "register" ? (
              "Create Account"
            ) : (
              "Play as Guest"
            )}
          </button>
        </form>
      </div>
    </div>
  );
}
