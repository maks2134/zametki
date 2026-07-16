import { getToken } from "./auth";
import type {
  ApiError,
  Category,
  Note,
  Room,
} from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function request<T>(
  path: string,
  options: RequestInit = {},
  auth = true,
): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");
  if (auth) {
    const token = getToken();
    if (token) headers.set("Authorization", `Bearer ${token}`);
  }

  const res = await fetch(`${API_URL}${path}`, { ...options, headers });
  if (res.status === 204) return undefined as T;

  const body = await res.json().catch(() => null);
  if (!res.ok) {
    const err = body as ApiError | null;
    throw new Error(err?.error?.message ?? `Request failed (${res.status})`);
  }
  return body as T;
}

export async function createRoom(input: {
  title: string;
  name: string;
  color: string;
}) {
  return request<{ room: Room; token: string; code: string }>(
    "/api/rooms",
    { method: "POST", body: JSON.stringify(input) },
    false,
  );
}

export async function joinRoom(input: {
  code: string;
  name: string;
  color: string;
}) {
  return request<{ room: Room; token: string }>(
    "/api/rooms/join",
    { method: "POST", body: JSON.stringify(input) },
    false,
  );
}

export async function getMyRoom() {
  return request<{ room: Room }>("/api/rooms/me");
}

export async function listNotes(params?: {
  category?: Category;
  limit?: number;
  before?: string;
}) {
  const q = new URLSearchParams();
  if (params?.category) q.set("category", params.category);
  if (params?.limit) q.set("limit", String(params.limit));
  if (params?.before) q.set("before", params.before);
  const qs = q.toString();
  return request<{ notes: Note[]; nextBefore?: string }>(
    `/api/notes${qs ? `?${qs}` : ""}`,
  );
}

export async function createNote(input: {
  title?: string;
  content: string;
  category: Category;
  color?: string;
  pinned?: boolean;
}) {
  return request<Note>("/api/notes", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateNote(
  id: string,
  input: {
    title?: string;
    content?: string;
    category?: Category;
    color?: string;
    pinned?: boolean;
  },
) {
  return request<Note>(`/api/notes/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export async function deleteNote(id: string) {
  return request<void>(`/api/notes/${id}`, { method: "DELETE" });
}

export async function addReaction(noteId: string, emoji: string) {
  return request<Note>(`/api/notes/${noteId}/reactions`, {
    method: "POST",
    body: JSON.stringify({ emoji }),
  });
}

export async function removeReaction(noteId: string) {
  return request<Note>(`/api/notes/${noteId}/reactions`, {
    method: "DELETE",
  });
}

export function getWsUrl(token: string) {
  const base = API_URL.replace(/^http/, "ws");
  return `${base}/ws?token=${encodeURIComponent(token)}`;
}
