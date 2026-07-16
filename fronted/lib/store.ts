"use client";

import { create } from "zustand";
import * as api from "./api";
import { getStoredMember, getStoredRoom, getToken, updateStoredRoom } from "./auth";
import type { ConnectionStatus } from "./ws";
import { RoomSocket } from "./ws";
import type { Category, Member, Note, Room, WSEvent } from "./types";

interface RoomState {
  room: Room | null;
  me: Member | null;
  notes: Note[];
  category: Category | "all";
  status: ConnectionStatus;
  loading: boolean;
  error: string | null;
  socket: RoomSocket | null;

  hydrate: () => void;
  setCategory: (c: Category | "all") => void;
  loadRoom: () => Promise<void>;
  loadNotes: () => Promise<void>;
  connectWs: () => void;
  disconnectWs: () => void;
  applyEvent: (e: WSEvent) => void;
  upsertNote: (n: Note) => void;
  removeNote: (id: string) => void;
}

function sortNotes(notes: Note[]) {
  return [...notes].sort((a, b) => {
    if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
    return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
  });
}

export const useRoomStore = create<RoomState>((set, get) => ({
  room: null,
  me: null,
  notes: [],
  category: "all",
  status: "offline",
  loading: false,
  error: null,
  socket: null,

  hydrate: () => {
    set({ room: getStoredRoom(), me: getStoredMember() });
  },

  setCategory: (category) => {
    set({ category });
    void get().loadNotes();
  },

  loadRoom: async () => {
    try {
      const { room } = await api.getMyRoom();
      const me = get().me ?? getStoredMember();
      updateStoredRoom(room);
      set({ room, me, error: null });
    } catch (e) {
      set({ error: e instanceof Error ? e.message : "Failed to load room" });
    }
  },

  loadNotes: async () => {
    set({ loading: true });
    try {
      const category = get().category;
      const { notes } = await api.listNotes(
        category === "all" ? undefined : { category },
      );
      set({ notes: sortNotes(notes), loading: false, error: null });
    } catch (e) {
      set({
        loading: false,
        error: e instanceof Error ? e.message : "Failed to load notes",
      });
    }
  },

  connectWs: () => {
    const token = getToken();
    if (!token) return;
    get().socket?.disconnect();
    const socket = new RoomSocket(
      token,
      (e) => get().applyEvent(e),
      (status) => set({ status }),
    );
    set({ socket });
    socket.connect();
  },

  disconnectWs: () => {
    get().socket?.disconnect();
    set({ socket: null, status: "offline" });
  },

  applyEvent: (e) => {
    switch (e.type) {
      case "note.created":
      case "note.updated":
      case "reaction.updated":
        get().upsertNote(e.data);
        break;
      case "note.deleted":
        get().removeNote(e.data.id);
        break;
      case "member.joined": {
        const room = get().room;
        if (!room) break;
        const exists = room.members.some((m) => m.id === e.data.id);
        const members = exists
          ? room.members
          : [...room.members, e.data];
        const next = { ...room, members };
        updateStoredRoom(next);
        set({ room: next });
        break;
      }
    }
  },

  upsertNote: (n) => {
    const category = get().category;
    if (category !== "all" && n.category !== category) {
      get().removeNote(n.id);
      return;
    }
    const notes = get().notes.filter((x) => x.id !== n.id);
    set({ notes: sortNotes([...notes, n]) });
  },

  removeNote: (id) => {
    set({ notes: get().notes.filter((n) => n.id !== id) });
  },
}));
