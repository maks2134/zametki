"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { AnimatePresence } from "framer-motion";
import { getToken } from "@/lib/auth";
import { useRoomStore } from "@/lib/store";
import { CategoryFilter } from "@/components/CategoryFilter";
import { NoteCard } from "@/components/NoteCard";
import { NoteComposer } from "@/components/NoteComposer";
import { RoomHeader } from "@/components/RoomHeader";

export default function RoomPage() {
  const router = useRouter();
  const {
    room,
    me,
    notes,
    category,
    status,
    loading,
    error,
    hydrate,
    loadRoom,
    loadNotes,
    connectWs,
    disconnectWs,
    setCategory,
    upsertNote,
    removeNote,
  } = useRoomStore();

  useEffect(() => {
    if (!getToken()) {
      router.replace("/");
      return;
    }
    hydrate();
    void loadRoom();
    void loadNotes();
    connectWs();
    return () => disconnectWs();
  }, [router, hydrate, loadRoom, loadNotes, connectWs, disconnectWs]);

  if (!room || !me) {
    return (
      <div className="flex min-h-dvh items-center justify-center bg-[#fff7f8] px-4 text-center text-rose-700">
        Загружаем ваше пространство…
      </div>
    );
  }

  return (
    <div className="min-h-dvh bg-[radial-gradient(ellipse_at_top,_#ffe0ea_0%,_transparent_50%),linear-gradient(180deg,_#fff7f8,_#fff)]">
      <RoomHeader room={room} status={status} />
      <main className="mx-auto max-w-3xl space-y-5 px-4 pt-5 pb-[calc(6.5rem+env(safe-area-inset-bottom))] sm:space-y-6 sm:py-8 sm:pb-28">
        <CategoryFilter value={category} onChange={setCategory} />
        {error && (
          <p className="rounded-2xl bg-rose-50 px-4 py-3 text-sm text-rose-700">
            {error}
          </p>
        )}
        {loading && notes.length === 0 ? (
          <p className="text-center text-muted-foreground">Загружаем идеи…</p>
        ) : notes.length === 0 ? (
          <div className="rounded-[1.75rem] border border-dashed border-rose-200 bg-white/50 px-5 py-12 text-center sm:rounded-[2rem] sm:px-6 sm:py-16">
            <p className="font-[family-name:var(--font-display)] text-2xl text-rose-900 sm:text-3xl">
              Пока тихо…
            </p>
            <p className="mt-2 text-sm text-rose-900/60 sm:text-base">
              Нажмите «Идея» внизу — партнёр увидит её мгновенно.
            </p>
          </div>
        ) : (
          <div className="space-y-3 sm:space-y-4">
            <AnimatePresence mode="popLayout">
              {notes.map((note) => (
                <NoteCard
                  key={note.id}
                  note={note}
                  members={room.members}
                  meId={me.id}
                  onUpdated={upsertNote}
                  onDeleted={removeNote}
                />
              ))}
            </AnimatePresence>
          </div>
        )}
      </main>
      <NoteComposer onCreated={upsertNote} variant="fab" />
    </div>
  );
}
