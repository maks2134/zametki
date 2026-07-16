import type { Member, Room } from "./types";

const TOKEN_KEY = "zametka_token";
const ROOM_KEY = "zametka_room";
const MEMBER_KEY = "zametka_member";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setSession(token: string, room: Room, member: Member) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(ROOM_KEY, JSON.stringify(room));
  localStorage.setItem(MEMBER_KEY, JSON.stringify(member));
}

export function getStoredRoom(): Room | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem(ROOM_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as Room;
  } catch {
    return null;
  }
}

export function getStoredMember(): Member | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem(MEMBER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as Member;
  } catch {
    return null;
  }
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(ROOM_KEY);
  localStorage.removeItem(MEMBER_KEY);
}

export function updateStoredRoom(room: Room) {
  localStorage.setItem(ROOM_KEY, JSON.stringify(room));
}
