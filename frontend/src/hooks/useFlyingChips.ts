import { useState, useEffect, useRef, useCallback } from "react";
import { GameStateResponse } from "../types";

export interface ChipParticle {
  id: number;
  startX: number;
  startY: number;
  endX: number;
  endY: number;
  delay: number;
}

const SEAT_COORDINATES = [
  { x: 50, y: 90 },
  { x: 15, y: 78 },
  { x: 8, y: 50 },
  { x: 15, y: 22 },
  { x: 50, y: 8 },
  { x: 85, y: 22 },
  { x: 92, y: 50 },
  { x: 85, y: 78 },
];

let chipIdCounter = 0;
const nextChipId = () => ++chipIdCounter;

export function useFlyingChips(gameState: GameStateResponse, playerId: string) {
  const [flyingChips, setFlyingChips] = useState<ChipParticle[]>([]);
  const prevGameStateRef = useRef<GameStateResponse | null>(null);

  const removeChip = useCallback((id: number) => {
    setFlyingChips((prev) => prev.filter((c) => c.id !== id));
  }, []);

  useEffect(() => {
    if (!gameState) return;

    if (!prevGameStateRef.current) {
      prevGameStateRef.current = gameState;
      return;
    }

    const prev = prevGameStateRef.current;
    const current = gameState;

    const hero = current.players.find((x) => x.id === playerId);
    const heroSeat = hero ? hero.seat : 0;
    const rotateSeat = (seatIdx: number) => {
      if (!hero) return seatIdx;
      return (seatIdx - heroSeat + 8) % 8;
    };

    const prevPlayersMap = new Map(prev.players.map((p) => [p.id, p]));

    current.players.forEach((p) => {
      const prevBet = prevPlayersMap.get(p.id)?.bet ?? 0;
      if (p.bet > prevBet) {
        const rotatedIndex = rotateSeat(p.seat);
        if (rotatedIndex >= 0 && rotatedIndex < 8) {
          const start = SEAT_COORDINATES[rotatedIndex] || { x: 50, y: 50 };
          const end = { x: 50, y: 32 };

          const count = 6;
          const newChips = Array.from({ length: count }).map((_, i) => ({
            id: nextChipId(),
            startX: start.x,
            startY: start.y,
            endX: end.x,
            endY: end.y,
            delay: i * 80,
          }));
          setFlyingChips((prevChips) => [...prevChips, ...newChips]);
        }
      }
    });

    const hadWinners = prev.last_winners && prev.last_winners.length > 0;
    if (current.last_winners && current.last_winners.length > 0 && !hadWinners && current.phase === "Waiting") {
      current.last_winners.forEach((w) => {
        const matchingPlayer = current.players.find((x) => x.id === w.player_id || x.name === w.player_name);
        if (matchingPlayer) {
          const rotatedIndex = rotateSeat(matchingPlayer.seat);
          if (rotatedIndex >= 0 && rotatedIndex < 8) {
            const start = { x: 50, y: 32 };
            const end = SEAT_COORDINATES[rotatedIndex] || { x: 50, y: 50 };

            const count = 12;
            const newChips = Array.from({ length: count }).map((_, i) => ({
              id: nextChipId(),
              startX: start.x,
              startY: start.y,
              endX: end.x,
              endY: end.y,
              delay: i * 60,
            }));
            setFlyingChips((prevChips) => [...prevChips, ...newChips]);
          }
        }
      });
    }

    prevGameStateRef.current = gameState;
  }, [gameState, playerId]);

  return {
    flyingChips,
    removeChip,
  };
}
