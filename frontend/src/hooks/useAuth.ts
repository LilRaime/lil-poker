import { useState, useEffect, useCallback } from "react";
import { pokerApi } from "../services/api";

export function useAuth() {
  const [playerName, setPlayerName] = useState<string>(() => localStorage.getItem("poker_name") || "");
  const [playerChips, setPlayerChips] = useState<number>(() => Number(localStorage.getItem("poker_chips")) || 1000);
  const [playerId, setPlayerId] = useState<string>(() => localStorage.getItem("poker_id") || "");

  const handleAuthSuccess = useCallback((uuid: string, username: string, chips: number, isGuest?: boolean, guestPassword?: string) => {
    localStorage.setItem("poker_id", uuid);
    localStorage.setItem("poker_name", username);
    localStorage.setItem("poker_chips", String(chips));
    if (isGuest) {
      localStorage.setItem("guest_uuid", uuid);
      if (guestPassword) {
        localStorage.setItem("guest_password", guestPassword);
      }
    }
    setPlayerId(uuid);
    setPlayerName(username);
    setPlayerChips(chips);
  }, []);

  const handleClearAuth = useCallback(() => {
    localStorage.removeItem("poker_id");
    localStorage.removeItem("poker_name");
    localStorage.removeItem("poker_chips");
    localStorage.removeItem("guest_uuid");
    localStorage.removeItem("guest_password");
    setPlayerId("");
    setPlayerName("");
  }, []);

  const checkAuth = useCallback(async () => {
    try {
      const data = await pokerApi.checkAuth();
      handleAuthSuccess(data.uuid, data.username, data.chips);
    } catch (err: any) {
      if (err.status === 401) {
        const guestUuid = localStorage.getItem("guest_uuid");
        const guestPassword = localStorage.getItem("guest_password");
        if (guestUuid && guestPassword) {
          try {
            const data = await pokerApi.resumeGuest(guestUuid, guestPassword);
            handleAuthSuccess(data.uuid, data.username, data.chips, true, guestPassword);
            return;
          } catch {
            localStorage.removeItem("guest_uuid");
            localStorage.removeItem("guest_password");
          }
        }
      }
      handleClearAuth();
    }
  }, [handleAuthSuccess, handleClearAuth]);

  const updateChips = useCallback((chips: number) => {
    setPlayerChips(chips);
    localStorage.setItem("poker_chips", String(chips));
  }, []);

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  return {
    playerId,
    playerName,
    playerChips,
    setPlayerId,
    setPlayerName,
    updateChips,
    handleAuthSuccess,
    handleClearAuth,
    checkAuth,
  };
}
