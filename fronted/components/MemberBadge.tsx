"use client";

import { cn } from "@/lib/utils";

export function MemberBadge({
  name,
  color,
  size = "md",
}: {
  name: string;
  color: string;
  size?: "sm" | "md";
}) {
  const initial = name.trim().charAt(0).toUpperCase() || "?";
  return (
    <span className="inline-flex items-center gap-2">
      <span
        className={cn(
          "inline-flex items-center justify-center rounded-full font-medium text-white shadow-sm",
          size === "sm" ? "h-7 w-7 text-xs" : "h-9 w-9 text-sm",
        )}
        style={{ backgroundColor: color }}
      >
        {initial}
      </span>
      <span className={cn("font-medium", size === "sm" ? "text-sm" : "text-base")}>
        {name}
      </span>
    </span>
  );
}
