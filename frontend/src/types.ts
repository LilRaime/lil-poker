export type Card = string;

export interface RoomInfo {
  id: string;
  name: string;
  player_count: number;
  max_players: number;
  phase: string;
  small_blind: number;
  big_blind: number;
  creator_id: string;
  blind_escalation_mins?: number;
}

export interface PlayerStatus {
  id: string;
  name: string;
  chips: number;
  hole?: Card[];
  bet: number;
  folded: boolean;
  all_in: boolean;
  acted: boolean;
  current_hand?: string;
  sitting_out: boolean;
  seat: number;
  reaction?: string;
  hands_played: number;
  hands_vpip: number;
  biggest_pot_won: number;
  is_small_blind?: boolean;
  is_big_blind?: boolean;
  exposed_cards?: boolean;
}

export interface WinnerStatus {
  player_id: string;
  player_name: string;
  hand_rank?: string;
  hand_cards?: Card[];
  amount: number;
}

export interface HandHistoryEntry {
  hand_num: number;
  board: Card[];
  winners: WinnerStatus[];
}

export interface ChatMessage {
  player_name: string;
  player_id?: string;
  text: string;
  time: number;
  system?: boolean;
}

export interface SubPot {
  amount: number;
  contributors: string[];
}

export interface GameStateResponse {
  players: PlayerStatus[];
  board: Card[];
  phase: string;
  pot: number;
  current_bet: number;
  active_idx: number;
  active_name?: string;
  active_player_id?: string;
  dealer_idx: number;
  small_blind: number;
  big_blind: number;
  last_winners?: WinnerStatus[];
  action_deadline: number;
  hand_count: number;
  chat_messages?: ChatMessage[];
  next_small_blind?: number;
  next_big_blind?: number;
  hands_until_raise?: number;
  observer_count: number;
  blinds_raise_deadline: number;
  hand_history?: HandHistoryEntry[];
  creator_id?: string;
  sub_pots?: SubPot[];
  observers?: string[];
}
