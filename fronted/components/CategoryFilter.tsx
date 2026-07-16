"use client";

import { CATEGORIES, type Category } from "@/lib/types";
import { cn } from "@/lib/utils";

export function CategoryFilter({
  value,
  onChange,
}: {
  value: Category | "all";
  onChange: (v: Category | "all") => void;
}) {
  return (
    <div className="-mx-4 overflow-x-auto overscroll-x-contain px-4 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
      <div className="flex w-max gap-2 pb-1">
        {CATEGORIES.map((c) => {
          const active = value === c.value;
          return (
            <button
              key={c.value}
              type="button"
              onClick={() => onChange(c.value)}
              className={cn(
                "h-10 shrink-0 snap-start rounded-full px-4 text-sm font-medium transition active:scale-[0.97]",
                active
                  ? "bg-rose-500 text-white shadow-sm shadow-rose-300/40"
                  : "bg-white/70 text-rose-950/70 ring-1 ring-rose-100",
              )}
            >
              {c.label}
            </button>
          );
        })}
      </div>
    </div>
  );
}
