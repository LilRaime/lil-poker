import React, { useCallback } from "react";
import { PlayerStatus, GameStateResponse } from "../types";
import { pokerApi } from "../services/api";
import { DEFAULT_SMALL_BLIND, DEFAULT_BIG_BLIND } from "../constants";

interface UseGameActionsOptions {
  playerId: string;
  roomId: string;
  hero: PlayerStatus | undefined;
  setGameState: (s: GameStateResponse | null) => void;
  setErrorMessage: (m: string | null) => void;
  updateChips: (chips: number) => void;
  handleClearAuth: () => void;
}

export interface UseGameActionsReturn {
  handleLeave: () => Promise<void>;
  handleJoin: (seatOrEvent?: number | React.SyntheticEvent) => Promise<void>;
  handleRebuy: () => Promise<void>;
  handleStartGame: () => Promise<void>;
  handleResetGame: () => Promise<void>;
  handleAction: (action: string, amount?: number) => Promise<void>;
  handleSitToggle: () => Promise<void>;
  handleSitInDirect: () => Promise<void>;
  handleRebuyAndSitIn: () => Promise<void>;
  handleStandUp: (targetUuid?: string) => Promise<void>;
  handleAddBot: () => Promise<void>;
  handleSetBlinds: (sb: number, bb: number) => Promise<void>;
  handleShowCardsToggle: (show: boolean) => Promise<void>;
}

export function useGameActions({
  playerId,
  roomId,
  hero,
  setGameState,
  setErrorMessage,
  updateChips,
  handleClearAuth,
}: UseGameActionsOptions): UseGameActionsReturn {
  const handleLeave = useCallback(async () => {
    if (hero) {
      try {
        await pokerApi.stand(roomId);
      } catch (err) {
        console.error("Failed to stand up on exit:", err);
      }
    }
    handleClearAuth();
    setGameState(null);
    try {
      await pokerApi.logout();
    } catch (err) {
      console.error("Logout request failed:", err);
    }
  }, [hero, roomId, handleClearAuth, setGameState]);

  const handleJoin = useCallback(
    async (seatOrEvent: any = -1) => {
      let seat = -1;
      if (typeof seatOrEvent === "number") {
        seat = seatOrEvent;
      } else if (seatOrEvent && typeof seatOrEvent === "object" && "preventDefault" in seatOrEvent) {
        seatOrEvent.preventDefault();
      }
      if (!playerId || !roomId) return;
      try {
        await pokerApi.joinTable(playerId, roomId, seat);
      } catch (err: any) {
        if (err.status === 401) {
          handleLeave();
          setErrorMessage("Session expired or invalid user. Please log in again.");
        } else {
          setErrorMessage(err.message);
        }
      }
    },
    [playerId, roomId, handleLeave, setErrorMessage]
  );

  const handleRebuy = useCallback(async () => {
    if (!playerId) return;
    try {
      const data = await pokerApi.rebuy(playerId, roomId);
      updateChips(data.chips);
    } catch (err: any) {
      setErrorMessage(err.message);
    }
  }, [playerId, roomId, updateChips, setErrorMessage]);

  const handleStartGame = useCallback(async () => {
    try {
      await pokerApi.startGame(roomId);
    } catch (err: any) {
      setErrorMessage(err.message);
    }
  }, [roomId, setErrorMessage]);

  const handleResetGame = useCallback(async () => {
    try {
      await pokerApi.resetGame(DEFAULT_SMALL_BLIND, DEFAULT_BIG_BLIND, roomId);
      setErrorMessage("Game reset. Waiting for players to join.");
    } catch (err: any) {
      setErrorMessage(err.message);
    }
  }, [roomId, setErrorMessage]);

  const handleAction = useCallback(
    async (action: string, amount: number = 0) => {
      try {
        await pokerApi.act(playerId, action, amount, roomId);
      } catch (err: any) {
        if (err.status === 401) {
          handleLeave();
          setErrorMessage("Session expired. Please log in again.");
        } else {
          setErrorMessage(err.message);
        }
      }
    },
    [playerId, roomId, handleLeave, setErrorMessage]
  );

  const handleSitToggle = useCallback(async () => {
    if (!hero) return;
    const nextAction = hero.sitting_out ? "in" : "out";
    try {
      await pokerApi.sit(nextAction, roomId);
    } catch (err: any) {
      setErrorMessage(err.message);
    }
  }, [hero, roomId, setErrorMessage]);

  const handleSitInDirect = useCallback(async () => {
    try {
      await pokerApi.sit("in", roomId);
    } catch (err: any) {
      setErrorMessage(err.message);
    }
  }, [roomId, setErrorMessage]);

  const handleRebuyAndSitIn = useCallback(async () => {
    if (!playerId) return;
    try {
      const data = await pokerApi.rebuy(playerId, roomId);
      updateChips(data.chips);
      await pokerApi.sit("in", roomId);
    } catch (err: any) {
      setErrorMessage(err.message);
    }
  }, [playerId, roomId, updateChips, setErrorMessage]);

  const handleStandUp = useCallback(async (targetUuid?: string) => {
    try {
      await pokerApi.stand(roomId, targetUuid);
    } catch (err: any) {
      if (err.status === 401) {
        handleLeave();
        setErrorMessage("Session expired or invalid user. Please log in again.");
      } else {
        setErrorMessage(err.message);
      }
    }
  }, [roomId, handleLeave, setErrorMessage]);

  const handleAddBot = useCallback(async () => {
    try {
      await pokerApi.addBot(roomId);
    } catch (err: any) {
      setErrorMessage(err.message);
    }
  }, [roomId, setErrorMessage]);

  const handleSetBlinds = useCallback(
    async (sb: number, bb: number) => {
      try {
        await pokerApi.setBlinds(sb, bb, roomId);
      } catch (err: any) {
        setErrorMessage(err.message);
      }
    },
    [roomId, setErrorMessage]
  );

  const handleShowCardsToggle = useCallback(
    async (show: boolean) => {
      try {
        await pokerApi.showCards(show, roomId);
      } catch (err: any) {
        setErrorMessage(err.message);
      }
    },
    [roomId, setErrorMessage]
  );

  return {
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
    handleAddBot,
    handleSetBlinds,
    handleShowCardsToggle,
  };
}
