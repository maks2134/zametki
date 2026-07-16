"use client";

import { useState } from "react";
import { Plus } from "lucide-react";
import { createNote } from "@/lib/api";
import type { Category, Note } from "@/lib/types";
import { CATEGORIES } from "@/lib/types";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

const NOTE_COLORS = ["#ffe5ec", "#fff0db", "#e8fff4", "#eef2ff", "#fff5f7"];

export function NoteComposer({
  onCreated,
  variant = "fab",
}: {
  onCreated: (n: Note) => void;
  variant?: "fab" | "inline";
}) {
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [category, setCategory] = useState<Category>("idea");
  const [color, setColor] = useState(NOTE_COLORS[0]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!content.trim()) {
      setError("Напишите хоть пару слов");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const note = await createNote({
        title: title.trim() || undefined,
        content: content.trim(),
        category,
        color,
      });
      onCreated(note);
      setTitle("");
      setContent("");
      setCategory("idea");
      setOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить");
    } finally {
      setBusy(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {variant === "fab" ? (
          <Button
            className={cn(
              "fixed z-30 h-14 rounded-full bg-rose-500 px-5 text-base text-white shadow-[0_12px_32px_-8px_rgba(225,70,100,0.65)] hover:bg-rose-600",
              "right-4 bottom-[calc(1rem+env(safe-area-inset-bottom))] sm:right-6 sm:bottom-6",
              "active:scale-[0.97]",
            )}
            aria-label="Новая идея"
          >
            <Plus className="mr-1 h-5 w-5" />
            Идея
          </Button>
        ) : (
          <Button className="h-11 rounded-full bg-rose-500 px-5 text-white shadow-lg shadow-rose-300/50 hover:bg-rose-600">
            <Plus className="mr-1 h-4 w-4" />
            Идея
          </Button>
        )}
      </DialogTrigger>
      <DialogContent
        className={cn(
          "flex max-h-[min(92dvh,720px)] w-[calc(100%-1.25rem)] max-w-lg flex-col gap-0 overflow-hidden border-none bg-[#fff8f9] p-0",
          "top-[max(1rem,env(safe-area-inset-top))] translate-y-0 sm:top-1/2 sm:-translate-y-1/2",
          "rounded-3xl sm:rounded-3xl",
        )}
      >
        <DialogHeader className="shrink-0 px-5 pt-5 pr-12 pb-3 sm:px-6 sm:pt-6">
          <DialogTitle className="font-[family-name:var(--font-display)] text-xl text-rose-950 sm:text-2xl">
            Новая идея для вас двоих
          </DialogTitle>
        </DialogHeader>
        <form
          onSubmit={(e) => void submit(e)}
          className="flex min-h-0 flex-1 flex-col"
        >
          <div className="min-h-0 flex-1 space-y-4 overflow-y-auto px-5 pb-4 sm:px-6">
            <div className="space-y-2">
              <Label htmlFor="title">Заголовок</Label>
              <Input
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Необязательно"
                className="h-11 rounded-xl bg-white text-base"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="content">Текст</Label>
              <Textarea
                id="content"
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="Куда сходить, что подарить, о чём мечтаете…"
                rows={4}
                className="rounded-xl bg-white text-base"
                required
              />
            </div>
            <div className="space-y-2">
              <Label>Категория</Label>
              <div className="-mx-1 flex gap-2 overflow-x-auto px-1 pb-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
                {CATEGORIES.filter((c) => c.value !== "all").map((c) => (
                  <Button
                    key={c.value}
                    type="button"
                    size="sm"
                    variant={category === c.value ? "default" : "outline"}
                    className={cn(
                      "h-10 shrink-0 rounded-full px-3",
                      category === c.value &&
                        "bg-rose-500 text-white hover:bg-rose-600",
                    )}
                    onClick={() => setCategory(c.value as Category)}
                  >
                    {c.label}
                  </Button>
                ))}
              </div>
            </div>
            <div className="space-y-2">
              <Label>Цвет карточки</Label>
              <div className="flex gap-3">
                {NOTE_COLORS.map((c) => (
                  <button
                    key={c}
                    type="button"
                    aria-label={c}
                    onClick={() => setColor(c)}
                    className="h-11 w-11 rounded-full border-2 transition active:scale-95"
                    style={{
                      backgroundColor: c,
                      borderColor: color === c ? "#e85d75" : "transparent",
                    }}
                  />
                ))}
              </div>
            </div>
            {error && <p className="text-sm text-rose-600">{error}</p>}
          </div>
          <div className="shrink-0 border-t border-rose-100/80 bg-[#fff8f9] px-5 py-4 pb-[max(1rem,env(safe-area-inset-bottom))] sm:px-6">
            <Button
              type="submit"
              disabled={busy}
              className="h-12 w-full rounded-full bg-rose-500 text-base text-white hover:bg-rose-600"
            >
              {busy ? "Сохраняем…" : "Поделиться"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
