import { useState, useCallback } from "react";

interface UseRoomReturn {
  currentRoomId: string | null;
  currentRoomCreatorId: string;
  enterRoom: (roomId: string, creatorId?: string) => void;
  leaveRoom: () => void;
  setRoomCreatorId: (creatorId: string) => void;
}

export function useRoom(): UseRoomReturn {
  const [currentRoomId, setCurrentRoomId] = useState<string | null>(() => {
    const urlRoom = new URLSearchParams(window.location.search).get("room");
    if (urlRoom) return urlRoom.toUpperCase();
    return localStorage.getItem("lastRoom");
  });
  const [currentRoomCreatorId, setCurrentRoomCreatorId] = useState<string>(() => {
    return localStorage.getItem("lastRoomCreatorId") ?? "";
  });

  const setRoomCreatorId = useCallback((creatorId: string) => {
    setCurrentRoomCreatorId(creatorId);
    if (creatorId) {
      localStorage.setItem("lastRoomCreatorId", creatorId);
    } else {
      localStorage.removeItem("lastRoomCreatorId");
    }
  }, []);

  const enterRoom = useCallback((roomId: string, creatorId?: string) => {
    setCurrentRoomId(roomId);
    setCurrentRoomCreatorId(creatorId ?? "");
    localStorage.setItem("lastRoom", roomId);
    if (creatorId) {
      localStorage.setItem("lastRoomCreatorId", creatorId);
    } else {
      localStorage.removeItem("lastRoomCreatorId");
    }
    history.pushState({}, "", `/?room=${encodeURIComponent(roomId)}`);
  }, []);

  const leaveRoom = useCallback(() => {
    setCurrentRoomId(null);
    setCurrentRoomCreatorId("");
    localStorage.removeItem("lastRoom");
    localStorage.removeItem("lastRoomCreatorId");
    history.pushState({}, "", "/");
  }, []);

  return { currentRoomId, currentRoomCreatorId, enterRoom, leaveRoom, setRoomCreatorId };
}
