import { useEffect, useMemo, useState } from "react";

import Lobby from "./components/Lobby";
import ActionPanel from "./components/ActionPanel";
import WinnersOverlay from "./components/WinnersOverlay";
import Auth from "./components/Auth";
import Header from "./components/Header";
import PokerTable from "./components/PokerTable";
import SitOutBanner from "./components/SitOutBanner";
import WaitingPanel from "./components/WaitingPanel";
import { usePokerSocket } from "./hooks/usePokerSocket";
import { useAuth } from "./hooks/useAuth";
import { useGameActions } from "./hooks/useGameActions";
import type { UseGameActionsReturn } from "./hooks/useGameActions";
import { useRoom } from "./hooks/useRoom";
import ChatPanel from "./components/ChatPanel";
import RoomBrowser from "./components/RoomBrowser";

export default function App() {
  const {
    playerId,
    playerName,
    playerChips,
    updateChips,
    handleAuthSuccess,
    handleClearAuth,
  } = useAuth();

  const { currentRoomId, currentRoomCreatorId, enterRoom, leaveRoom, setRoomCreatorId } = useRoom();

  const { gameState, setGameState, errorMessage, setErrorMessage, connectionState, sendMessage } =
    usePokerSocket(playerId, currentRoomId ?? "", () => {
      setErrorMessage("The room was closed or the server restarted.");
      leaveRoom();
      setGameState(null);
    });

  useEffect(() => {
    if (gameState?.creator_id && gameState.creator_id !== currentRoomCreatorId) {
      setRoomCreatorId(gameState.creator_id);
    }
  }, [gameState?.creator_id, currentRoomCreatorId, setRoomCreatorId]);

  const [autoRebuy, setAutoRebuy] = useState(() => {
    return localStorage.getItem("auto_rebuy") === "true";
  });

  useEffect(() => {
    localStorage.setItem("auto_rebuy", String(autoRebuy));
  }, [autoRebuy]);

  const [rebuying, setRebuying] = useState(false);

  const handleFullLeave = () => {
    leaveRoom();
    setGameState(null);
    handleClearAuth();
  };

  useEffect(() => {
    if (!playerId) {
      leaveRoom();
      setGameState(null);
    }
  }, [playerId, leaveRoom, setGameState]);

  useEffect(() => {
    let timer: ReturnType<typeof setTimeout>;
    if (gameState?.phase === "Closed") {
      setErrorMessage("The host has closed the room. Returning to lobby...");
      timer = setTimeout(() => {
        leaveRoom();
        setGameState(null);
      }, 3000);
    }
    return () => {
      if (timer) clearTimeout(timer);
    };
  }, [gameState?.phase, leaveRoom, setGameState, setErrorMessage]);

  useEffect(() => {
    if (errorMessage === "collision") {
      setErrorMessage("Disconnected: You have opened this game in another tab/device.");
      leaveRoom();
      setGameState(null);
    }
  }, [errorMessage, leaveRoom, setGameState, setErrorMessage]);

  useEffect(() => {
    if (errorMessage) {
      const timer = setTimeout(() => setErrorMessage(null), 5000);
      return () => clearTimeout(timer);
    }
    return;
  }, [errorMessage, setErrorMessage]);

  const hero = gameState?.players.find((p) => p.id === playerId);

  useEffect(() => {
    if (hero) updateChips(hero.chips);
  }, [hero, updateChips]);

  const {
    handleLeave,
    handleJoin,
    handleRebuy,
    handleStartGame,
    handleResetGame,
    handleAction,
    handleSitToggle,
    handleSitInDirect,
    handleRebuyAndSitIn,
    handleStandUp,
    handleSetBlinds,
    handleShowCardsToggle,
  } = useGameActions({
    playerId,
    roomId: currentRoomId ?? "",
    hero,
    setGameState,
    setErrorMessage,
    updateChips,
    handleClearAuth: handleFullLeave,
  }) satisfies UseGameActionsReturn;

  useEffect(() => {
    if (autoRebuy && hero && hero.chips === 0 && hero.sitting_out && !rebuying) {
      setRebuying(true);
      handleRebuyAndSitIn().finally(() => {
        setRebuying(false);
      });
    }
  }, [autoRebuy, hero, rebuying, handleRebuyAndSitIn]);


  const isShowdown = !!(
    gameState?.phase === "Waiting" &&
    gameState?.last_winners?.length
  );

  const winningCards = useMemo(() => {
    if (!isShowdown || !gameState?.last_winners) return [];
    return gameState.last_winners.flatMap((w) => w.hand_cards ?? []);
  }, [isShowdown, gameState?.last_winners]);

  if (!playerId) {
    return <Auth onAuthSuccess={handleAuthSuccess} />;
  }

  if (!currentRoomId) {
    return (
      <RoomBrowser
        playerName={playerName}
        playerChips={playerChips}
        onJoinRoom={enterRoom}
        onLogout={handleFullLeave}
      />
    );
  }

  return (
    <div className="flex flex-col min-h-screen bg-slate-950 text-slate-100 font-sans">
      {errorMessage && (
        <div className="fixed top-4 inset-x-0 flex justify-center z-50 pointer-events-none">
          <div className="pointer-events-auto bg-red-950/80 border border-red-500/30 text-red-200 px-6 py-3 rounded-full shadow-2xl backdrop-blur flex items-center space-x-3 animate-bounce">
            <span className="text-red-500 font-bold">⚠️</span>
            <span className="text-sm font-semibold">{errorMessage}</span>
            <button
              onClick={() => setErrorMessage(null)}
              className="ml-1 text-red-400 hover:text-red-200 font-bold text-lg leading-none transition-colors"
              aria-label="Dismiss error"
            >
              ×
            </button>
          </div>
        </div>
      )}

      {playerId && connectionState === "disconnected" && !errorMessage && (
        <div className="fixed top-4 left-1/2 -translate-x-1/2 z-50 bg-slate-800/90 border border-amber-500/40 text-amber-200 px-6 py-3 rounded-full shadow-2xl backdrop-blur flex items-center space-x-2">
          <span className="animate-spin text-amber-400">⟳</span>
          <span className="text-sm font-semibold">Reconnecting to server...</span>
        </div>
      )}

      <Header
        gameState={gameState}
        isAuthenticated={!!playerId}
        playerName={playerName}
        hero={hero}
        connectionState={connectionState}
        onSitToggle={handleSitToggle}
        onLeave={handleLeave}
        autoRebuy={autoRebuy}
        setAutoRebuy={setAutoRebuy}
        currentRoomId={currentRoomId}
        onLeaveRoom={() => {
          if (hero) {
            handleStandUp().catch(console.error);
          }
          leaveRoom();
          setGameState(null);
        }}
        onStandUp={handleStandUp}
      />

      <main className="flex-1 flex flex-col items-center justify-center p-4 max-w-7xl w-full mx-auto relative pb-20 sm:pb-4">
        {!hero ? (
          <Lobby
            playerName={playerName}
            playerChips={playerChips}
            onJoin={handleJoin}
            onReset={handleResetGame}
            onLogout={handleLeave}
            onRebuy={handleRebuy}
          />
        ) : (
          <div className="w-full grid grid-cols-1 lg:grid-cols-4 gap-6 items-start my-1 sm:my-1.5">
            <div className="lg:col-span-3 flex flex-col items-center">
              {hero.sitting_out && (
                <SitOutBanner
                  hero={hero}
                  onSitIn={handleSitInDirect}
                  onRebuyAndSitIn={handleRebuyAndSitIn}
                />
              )}

              {gameState?.phase === "Waiting" && (
                <WaitingPanel
                  gameState={gameState}
                  isCreator={!!playerId && playerId === currentRoomCreatorId}
                  onStart={handleStartGame}
                  onReset={handleResetGame}
                  onSetBlinds={handleSetBlinds}
                />
              )}

              {gameState && (
                <PokerTable
                  gameState={gameState}
                  playerId={playerId}
                  onJoin={handleJoin}
                  winningCards={winningCards}
                  isShowdown={isShowdown}
                />
              )}

              {gameState?.phase !== "Waiting" && gameState && (
                <div className="w-full max-w-2xl mt-8 sm:mt-8 glass-panel-heavy p-4 rounded-2xl border border-white/10 shadow-2xl flex flex-col space-y-4">
                  <ActionPanel
                    gameState={gameState}
                    playerId={playerId}
                    onAction={handleAction}
                  />
                  {hero && !hero.sitting_out && (
                    <div className="flex items-center justify-between border-t border-white/5 pt-3">
                      <span className="text-xs text-slate-400 font-semibold">Expose cards to the table?</span>
                      <label className="flex items-center space-x-1.5 text-xs font-semibold text-slate-300 cursor-pointer select-none bg-slate-800 border border-white/5 px-3 py-1.5 rounded-lg hover:bg-slate-700 transition-colors">
                        <input
                          type="checkbox"
                          checked={hero.exposed_cards}
                          onChange={(e) => handleShowCardsToggle(e.target.checked)}
                          className="accent-purple-500 rounded"
                        />
                        <span>Show Cards</span>
                      </label>
                    </div>
                  )}
                </div>
              )}

              {isShowdown && (
                <WinnersOverlay
                  lastWinners={gameState!.last_winners!}
                  onStartGame={handleStartGame}
                  onResetGame={handleResetGame}
                  isCreator={!!playerId && playerId === currentRoomCreatorId}
                />
              )}
            </div>

            <div className="lg:col-span-1 w-full lg:sticky lg:top-24">
              <ChatPanel
                chatMessages={gameState?.chat_messages || []}
                playerId={playerId}
                onSendMessage={sendMessage}
                handHistory={gameState?.hand_history || []}
                observers={gameState?.observers || []}
              />
            </div>
          </div>
        )}
      </main>
      {hero && (
        <div className="sm:hidden fixed bottom-0 left-0 right-0 z-40 glass-panel-heavy border-t border-white/10 px-3 py-2 flex items-center justify-between gap-2">
          <div className="flex items-center gap-1 text-xs font-black text-amber-400 bg-slate-800/60 border border-white/5 px-2.5 py-1.5 rounded-lg whitespace-nowrap">
            <span className="font-mono">{hero.chips.toLocaleString()}</span>
            <span>🪙</span>
          </div>
          <label className="flex items-center space-x-1.5 text-xs font-semibold text-slate-300 cursor-pointer select-none bg-slate-800 border border-white/5 px-2.5 py-1.5 rounded-lg">
            <input
              type="checkbox"
              checked={autoRebuy}
              onChange={(e) => setAutoRebuy(e.target.checked)}
              className="accent-purple-500 rounded"
            />
            <span>Auto</span>
          </label>
          <button
            onClick={handleSitToggle}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition-all ${hero.sitting_out
              ? "bg-amber-600/20 text-amber-300 border-amber-500/20"
              : "bg-slate-800 text-slate-300 border-white/5"
              }`}
          >
            {hero.sitting_out ? "Sit In" : "Sit Out"}
          </button>
          <button
            onClick={() => handleStandUp().catch(console.error)}
            className="px-3 py-1.5 rounded-lg text-xs font-semibold border bg-red-950/20 text-red-400 border-red-500/10 transition-all"
          >
            Stand Up
          </button>
          <button
            onClick={() => {
              handleStandUp().catch(console.error);
              leaveRoom();
              setGameState(null);
            }}
            className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-red-900/30 text-red-300 border border-red-500/20 transition-all"
          >
            Leave
          </button>
        </div>
      )}
    </div>
  );
}
