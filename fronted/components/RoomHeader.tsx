"use client";

import { useState } from "react";
import { Copy, Heart, Check } from "lucide-react";
import type { Note, Room } from "@/lib/types";
import { MemberBadge } from "@/components/MemberBadge";
import { ConnectionStatusBadge } from "@/components/ConnectionStatus";
import type { ConnectionStatus } from "@/lib/ws";
import { Button } from "@/components/ui/button";

export function RoomHeader({
  room,
  status,
}: {
  room: Room;
  status: ConnectionStatus;
  onCreated?: (n: Note) => void;
}) {
  const [copied, setCopied] = useState(false);

  async function copyInvite() {
    const url = `${window.location.origin}/join/${room.code}`;
    await navigator.clipboard.writeText(url);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1800);
  }

  return (
    <header className="sticky top-0 z-20 border-b border-rose-100/80 bg-[#fff5f7]/90 pt-[env(safe-area-inset-top)] backdrop-blur-xl">
      <div className="mx-auto flex max-w-3xl items-start justify-between gap-3 px-4 py-3 sm:items-center sm:py-4">
        <div className="min-w-0 flex-1 space-y-2">
          <div className="flex items-center gap-2">
            <Heart className="h-4 w-4 shrink-0 fill-rose-400 text-rose-400 sm:h-5 sm:w-5" />
            <h1 className="truncate font-[family-name:var(--font-display)] text-xl tracking-tight text-rose-950 sm:text-2xl">
              {room.title || "Заметка"}
            </h1>
          </div>
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1.5">
            {room.members.map((m) => (
              <MemberBadge key={m.id} name={m.name} color={m.color} size="sm" />
            ))}
            <ConnectionStatusBadge status={status} />
          </div>
        </div>

        <Button
          type="button"
          variant="outline"
          className="h-11 shrink-0 rounded-full border-rose-200 bg-white/80 px-3 sm:px-4"
          onClick={() => void copyInvite()}
          aria-label="Скопировать ссылку-приглашение"
        >
          {copied ? (
            <Check className="h-4 w-4 text-emerald-600" />
          ) : (
            <Copy className="h-4 w-4" />
          )}
          <span className="ml-1.5 hidden font-mono text-xs tracking-wider sm:inline">
            {copied ? "Скопировано" : room.code}
          </span>
        </Button>
      </div>
    </header>
  );
}
