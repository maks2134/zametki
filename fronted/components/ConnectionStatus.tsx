"use client";

import type { ConnectionStatus } from "@/lib/ws";
import { cn } from "@/lib/utils";

export function ConnectionStatusBadge({ status }: { status: ConnectionStatus }) {
  const label =
    status === "online"
      ? "Онлайн"
      : status === "connecting"
        ? "Подключение…"
        : "Офлайн";

  return (
    <span className="inline-flex items-center gap-2 text-xs text-muted-foreground">
      <span
        className={cn(
          "h-2 w-2 rounded-full",
          status === "online" && "bg-emerald-500",
          status === "connecting" && "bg-amber-400 animate-pulse",
          status === "offline" && "bg-rose-300",
        )}
      />
      {label}
    </span>
  );
}
