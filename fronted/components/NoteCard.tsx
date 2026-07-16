"use client";

import { motion } from "framer-motion";
import { Pin, Trash2 } from "lucide-react";
import { deleteNote, updateNote } from "@/lib/api";
import type { Member, Note } from "@/lib/types";
import { CATEGORIES } from "@/lib/types";
import { MemberBadge } from "@/components/MemberBadge";
import { ReactionBar } from "@/components/ReactionBar";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";

export function NoteCard({
  note,
  members,
  meId,
  onUpdated,
  onDeleted,
}: {
  note: Note;
  members: Member[];
  meId: string;
  onUpdated: (n: Note) => void;
  onDeleted: (id: string) => void;
}) {
  const author = members.find((m) => m.id === note.authorId);
  const isMine = note.authorId === meId;
  const categoryLabel =
    CATEGORIES.find((c) => c.value === note.category)?.label ?? note.category;

  async function togglePin() {
    onUpdated(await updateNote(note.id, { pinned: !note.pinned }));
  }

  async function remove() {
    if (!window.confirm("Удалить эту идею?")) return;
    await deleteNote(note.id);
    onDeleted(note.id);
  }

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 16, scale: 0.98 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, scale: 0.96 }}
      transition={{ type: "spring", stiffness: 320, damping: 28 }}
    >
      <Card
        className="overflow-hidden border-none bg-white/85 shadow-[0_10px_40px_-20px_rgba(190,60,90,0.35)] backdrop-blur"
        style={{
          borderTop: `3px solid ${note.color || author?.color || "#e85d75"}`,
        }}
      >
        <CardHeader className="flex flex-row items-start justify-between gap-2 space-y-0 px-4 pt-4 pb-2 sm:px-6 sm:pt-6">
          <div className="min-w-0 flex-1 space-y-1.5">
            {author && (
              <MemberBadge name={author.name} color={author.color} size="sm" />
            )}
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
              <span>{categoryLabel}</span>
              <span className="hidden sm:inline">·</span>
              <span>
                {new Date(note.createdAt).toLocaleString("ru-RU", {
                  day: "numeric",
                  month: "short",
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </span>
              {note.pinned && (
                <span className="inline-flex items-center gap-1 text-rose-500">
                  <Pin className="h-3 w-3" /> избранное
                </span>
              )}
            </div>
          </div>
          {isMine && (
            <div className="-mr-1 flex shrink-0 gap-0.5">
              <Button
                type="button"
                size="icon"
                variant="ghost"
                className="h-11 w-11"
                onClick={() => void togglePin()}
                aria-label="Закрепить"
              >
                <Pin
                  className={
                    note.pinned ? "h-5 w-5 fill-rose-500 text-rose-500" : "h-5 w-5"
                  }
                />
              </Button>
              <Button
                type="button"
                size="icon"
                variant="ghost"
                className="h-11 w-11"
                onClick={() => void remove()}
                aria-label="Удалить"
              >
                <Trash2 className="h-5 w-5 text-rose-400" />
              </Button>
            </div>
          )}
        </CardHeader>
        <CardContent className="space-y-4 px-4 pb-4 sm:px-6 sm:pb-6">
          {note.title ? (
            <h3 className="font-[family-name:var(--font-display)] text-lg text-rose-950 sm:text-xl">
              {note.title}
            </h3>
          ) : null}
          <p className="whitespace-pre-wrap text-[15px] leading-relaxed break-words text-rose-950/80">
            {note.content}
          </p>
          <ReactionBar note={note} memberId={meId} onUpdated={onUpdated} />
        </CardContent>
      </Card>
    </motion.div>
  );
}
