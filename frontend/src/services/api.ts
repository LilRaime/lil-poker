import { RoomInfo } from "../types";

async function fetchJson<T = any>(url: string, method: string = "GET", body?: any): Promise<T> {
  const options: RequestInit = {
    method,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
  };
  if (body !== undefined) {
    options.body = JSON.stringify(body);
  }
  const res = await fetch(url, options);
  if (!res.ok) {
    let errMsg = "Request failed";
    try {
      const errData = await res.json();
      errMsg = errData.error || errMsg;
    } catch {
    }
    const error = new Error(errMsg) as any;
    error.status = res.status;
    throw error;
  }
  return res.json();
}

function roomUrl(path: string, roomId: string): string {
  const sep = path.includes("?") ? "&" : "?";
  return `${path}${sep}room=${encodeURIComponent(roomId)}`;
}

export interface AuthUserResponse {
  uuid: string;
  username: string;
  chips: number;
}

export interface JoinTableResponse {
  id: string;
  name: string;
  chips: number;
}

export interface RebuyResponse {
  uuid: string;
  username: string;
  chips: number;
}

export interface MessageResponse {
  message: string;
}

export const pokerApi = {
  async checkAuth(): Promise<AuthUserResponse> {
    return fetchJson<AuthUserResponse>("/api/auth/me");
  },

  async logout(): Promise<MessageResponse> {
    return fetchJson<MessageResponse>("/api/auth/logout", "POST");
  },

  async resumeGuest(uuid: string, password?: string): Promise<AuthUserResponse> {
    return fetchJson<AuthUserResponse>("/api/auth/guest/resume", "POST", { uuid, password });
  },

  async rebuy(uuid: string, roomId?: string): Promise<RebuyResponse> {
    const url = roomId ? roomUrl("/api/auth/rebuy", roomId) : "/api/auth/rebuy";
    return fetchJson<RebuyResponse>(url, "POST", { uuid });
  },

  async listRooms(): Promise<RoomInfo[]> {
    return fetchJson<RoomInfo[]>("/api/rooms");
  },

  async createRoom(name: string, smallBlind: number, bigBlind: number, blindEscalationMins?: number): Promise<RoomInfo> {
    return fetchJson<RoomInfo>("/api/rooms", "POST", {
      name,
      small_blind: smallBlind,
      big_blind: bigBlind,
      blind_escalation_mins: blindEscalationMins,
    });
  },

  async joinTable(uuid: string, roomId: string, seat: number = -1): Promise<JoinTableResponse> {
    return fetchJson<JoinTableResponse>(roomUrl("/api/game/players", roomId), "POST", { uuid, seat });
  },

  async startGame(roomId: string): Promise<any> {
    return fetchJson(roomUrl("/api/game/start", roomId), "POST");
  },

  async resetGame(smallBlind: number, bigBlind: number, roomId: string): Promise<MessageResponse> {
    return fetchJson<MessageResponse>(roomUrl("/api/game/create", roomId), "POST", {
      small_blind: smallBlind,
      big_blind: bigBlind,
    });
  },

  async act(playerId: string, action: string, amount: number = 0, roomId: string): Promise<any> {
    return fetchJson(roomUrl("/api/game/act", roomId), "POST", {
      player_id: playerId,
      action,
      amount,
    });
  },

  async sit(action: "in" | "out", roomId: string): Promise<MessageResponse> {
    return fetchJson<MessageResponse>(roomUrl("/api/game/sit", roomId), "POST", { action });
  },

  async stand(roomId: string): Promise<MessageResponse> {
    return fetchJson<MessageResponse>(roomUrl("/api/game/stand", roomId), "POST");
  },

  async setBlinds(smallBlind: number, bigBlind: number, roomId: string): Promise<MessageResponse> {
    return fetchJson<MessageResponse>(roomUrl("/api/game/blinds", roomId), "POST", {
      small_blind: smallBlind,
      big_blind: bigBlind,
    });
  },

  async showCards(show: boolean, roomId: string): Promise<MessageResponse> {
    return fetchJson<MessageResponse>(roomUrl("/api/game/show-cards", roomId), "POST", { show });
  },
};
