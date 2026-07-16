"use client";

import { useState } from "react";
import { addReaction, removeReaction } from "@/lib/api";
import type { Note } from "@/lib/types";
import { cn } from "@/lib/utils";

const EMOJIS = ["❤️", "🥰", "🔥", "✨", "😂", "👍"];

export function ReactionBar({
  note,
  memberId,
  onUpdated,
}: {
  note: Note;
  memberId: string;
  onUpdated: (n: Note) => void;
}) {
  const [busy, setBusy] = useState(false);
  const mine = note.reactions.find((r) => r.memberId === memberId);

  async function toggle(emoji: string) {
    if (busy) return;
    setBusy(true);
    try {
      if (mine?.emoji === emoji) {
        onUpdated(await removeReaction(note.id));
      } else {
        onUpdated(await addReaction(note.id, emoji));
      }
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="-mx-1 flex gap-1 overflow-x-auto px-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
      {EMOJIS.map((emoji) => {
        const count = note.reactions.filter((r) => r.emoji === emoji).length;
        const active = mine?.emoji === emoji;
        return (
          <button
            key={emoji}
            type="button"
            disabled={busy}
            onClick={() => void toggle(emoji)}
            className={cn(
              "inline-flex h-11 min-w-11 shrink-0 items-center justify-center gap-1 rounded-full px-2.5 text-base transition active:scale-95 disabled:opacity-50",
              active
                ? "bg-rose-100 ring-1 ring-rose-300"
                : "bg-rose-50/60 hover:bg-rose-50",
            )}
            aria-label={`Реакция ${emoji}`}
          >
            <span>{emoji}</span>
            {count > 0 && (
              <span className="text-xs text-muted-foreground">{count}</span>
            )}
          </button>
        );
      })}
    </div>
  );
}
