import { useState, useEffect, useRef, useCallback } from "react";
import { GameStateResponse } from "../types";
import { pokerApi } from "../services/api";

export type ConnectionState = "connecting" | "connected" | "disconnected";

export function usePokerSocket(playerId: string, roomId: string, onRoomNotFound?: () => void) {
  const [gameState, setGameState] = useState<GameStateResponse | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [connectionState, setConnectionState] = useState<ConnectionState>("disconnected");
  const wsRef = useRef<WebSocket | null>(null);
  const onRoomNotFoundRef = useRef(onRoomNotFound);

  useEffect(() => {
    onRoomNotFoundRef.current = onRoomNotFound;
  }, [onRoomNotFound]);

  useEffect(() => {
    if (!playerId || !roomId) {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      setConnectionState("disconnected");
      return;
    }

    let active = true;
    let socket: WebSocket | null = null;
    let reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;
    let currentDelay = 1000;
    const maxDelay = 16000;

    const connect = () => {
      if (!active) return;
      console.log(`Connecting to WebSocket for room ${roomId}...`);
      setConnectionState("connecting");
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      socket = new WebSocket(
        `${protocol}//${window.location.host}/api/game/ws?room=${encodeURIComponent(roomId)}`
      );

      socket.onopen = () => {
        if (active) {
          setErrorMessage(null);
          setConnectionState("connected");
          currentDelay = 1000;
        }
      };

      socket.onmessage = (event) => {
        if (!active) return;
        try {
          const state = JSON.parse(event.data) as GameStateResponse;
          setGameState(state);
        } catch (err) {
          console.error("Error parsing WebSocket message:", err);
        }
      };

      socket.onerror = (err) => {
        console.error("WebSocket error:", err);
        if (active) {
          setConnectionState("disconnected");
        }
      };

      socket.onclose = (e) => {
        if (!active) return;
        setConnectionState("disconnected");

        const isIntentional = e.code === 1000 || e.code >= 4000;
        if (!isIntentional) {
          pokerApi.listRooms()
            .then((rooms) => {
              if (!active) return;
              const roomExists = rooms.some((r) => r.id === roomId);
              if (!roomExists) {
                console.log(`Room ${roomId} no longer exists. Triggering onRoomNotFound.`);
                if (onRoomNotFoundRef.current) {
                  onRoomNotFoundRef.current();
                }
              } else {
                reconnectTimeoutId = setTimeout(() => {
                  currentDelay = Math.min(currentDelay * 2, maxDelay);
                  connect();
                }, currentDelay);
              }
            })
            .catch(() => {
              if (!active) return;
              reconnectTimeoutId = setTimeout(() => {
                currentDelay = Math.min(currentDelay * 2, maxDelay);
                connect();
              }, currentDelay);
            });
        }
      };

      wsRef.current = socket;
    };

    connect();

    return () => {
      active = false;
      if (reconnectTimeoutId) {
        clearTimeout(reconnectTimeoutId);
      }
      if (socket) {
        socket.close();
      }
      setConnectionState("disconnected");
    };
  }, [playerId, roomId]);

  const sendMessage = useCallback((text: string) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: "chat", text }));
    }
  }, []);

  return {
    gameState,
    setGameState,
    errorMessage,
    setErrorMessage,
    connectionState,
    sendMessage,
  };
}
