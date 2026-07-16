"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { Heart, Sparkles } from "lucide-react";
import { createRoom, joinRoom } from "@/lib/api";
import { getToken, setSession } from "@/lib/auth";
import { MEMBER_COLORS } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function HomePage() {
  const router = useRouter();
  const [title, setTitle] = useState("Наши заметки");
  const [name, setName] = useState("");
  const [color, setColor] = useState(MEMBER_COLORS[0]);
  const [code, setCode] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (getToken()) router.replace("/room");
  }, [router]);

  const hearts = useMemo(
    () =>
      Array.from({ length: 6 }, (_, i) => ({
        id: i,
        left: `${10 + i * 14}%`,
        delay: i * 0.4,
        size: 10 + (i % 3) * 5,
      })),
    [],
  );

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const res = await createRoom({
        title: title.trim() || "Наши заметки",
        name: name.trim(),
        color,
      });
      const me = res.room.members[0];
      setSession(res.token, res.room, me);
      router.push("/room");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка создания");
    } finally {
      setBusy(false);
    }
  }

  async function onJoin(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const res = await joinRoom({
        code: code.trim().toUpperCase(),
        name: name.trim(),
        color,
      });
      const me =
        res.room.members.find((m) => m.name === name.trim()) ??
        res.room.members[res.room.members.length - 1];
      setSession(res.token, res.room, me);
      router.push("/room");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось войти");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="relative min-h-dvh overflow-x-hidden">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top,_#ffd6e0_0%,_transparent_55%),radial-gradient(ellipse_at_bottom_right,_#ffe8cc_0%,_transparent_45%),linear-gradient(180deg,_#fff7f8_0%,_#fff_70%)]" />
      {hearts.map((h) => (
        <motion.span
          key={h.id}
          className="pointer-events-none absolute hidden text-rose-300/40 sm:block"
          style={{ left: h.left, bottom: "-5%", fontSize: h.size }}
          animate={{ y: ["0vh", "-110vh"], opacity: [0, 0.7, 0] }}
          transition={{
            duration: 12 + h.id,
            repeat: Infinity,
            delay: h.delay,
            ease: "linear",
          }}
        >
          ♥
        </motion.span>
      ))}

      <div className="relative mx-auto flex min-h-dvh max-w-xl flex-col justify-center px-4 py-10 pt-[max(2.5rem,env(safe-area-inset-top))] pb-[max(2.5rem,env(safe-area-inset-bottom))] sm:py-16">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7, ease: [0.22, 1, 0.36, 1] }}
          className="mb-7 text-center sm:mb-10"
        >
          <div className="mb-3 inline-flex items-center gap-2 rounded-full bg-white/70 px-3.5 py-1.5 text-xs text-rose-600 shadow-sm backdrop-blur sm:mb-4 sm:text-sm">
            <Sparkles className="h-3.5 w-3.5 sm:h-4 sm:w-4" />
            для двоих
          </div>
          <h1 className="font-[family-name:var(--font-display)] text-5xl tracking-tight text-rose-950 sm:text-7xl">
            Zametka
          </h1>
          <p className="mx-auto mt-3 max-w-md text-[15px] leading-relaxed text-rose-950/65 sm:mt-4 sm:text-lg">
            Общее пространство идей, свиданий и нежных мыслей — только для вас
            двоих.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 18 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.15, duration: 0.6 }}
          className="rounded-[1.75rem] border border-white/70 bg-white/80 p-5 shadow-[0_30px_80px_-40px_rgba(190,60,90,0.55)] backdrop-blur-xl sm:rounded-[2rem] sm:p-8"
        >
          <Tabs defaultValue="create">
            <TabsList className="mb-5 grid h-11 w-full grid-cols-2 bg-rose-50 sm:mb-6">
              <TabsTrigger value="create" className="text-sm">
                Создать
              </TabsTrigger>
              <TabsTrigger value="join" className="text-sm">
                Войти по коду
              </TabsTrigger>
            </TabsList>

            <div className="mb-5 space-y-3">
              <Label htmlFor="name">Ваше имя</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Как вас зовут?"
                className="h-12 rounded-xl bg-white text-base"
                autoComplete="nickname"
                required
              />
              <Label>Цвет</Label>
              <div className="flex flex-wrap gap-3">
                {MEMBER_COLORS.map((c) => (
                  <button
                    key={c}
                    type="button"
                    aria-label={c}
                    onClick={() => setColor(c)}
                    className="h-11 w-11 rounded-full border-2 transition active:scale-95"
                    style={{
                      backgroundColor: c,
                      borderColor: color === c ? "#1f0a12" : "transparent",
                      transform: color === c ? "scale(1.06)" : undefined,
                    }}
                  />
                ))}
              </div>
            </div>

            <TabsContent value="create">
              <form onSubmit={(e) => void onCreate(e)} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="title">Название пространства</Label>
                  <Input
                    id="title"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                    className="h-12 rounded-xl bg-white text-base"
                  />
                </div>
                <Button
                  type="submit"
                  disabled={busy || !name.trim()}
                  className="h-12 w-full rounded-full bg-rose-500 text-base text-white hover:bg-rose-600"
                >
                  <Heart className="mr-2 h-4 w-4 fill-white" />
                  {busy ? "Создаём…" : "Создать пространство"}
                </Button>
              </form>
            </TabsContent>

            <TabsContent value="join">
              <form onSubmit={(e) => void onJoin(e)} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="code">Код приглашения</Label>
                  <Input
                    id="code"
                    value={code}
                    onChange={(e) => setCode(e.target.value.toUpperCase())}
                    placeholder="K7M4QP"
                    className="h-12 rounded-xl bg-white text-base tracking-[0.2em]"
                    autoCapitalize="characters"
                    autoCorrect="off"
                    required
                  />
                </div>
                <Button
                  type="submit"
                  disabled={busy || !name.trim() || !code.trim()}
                  className="h-12 w-full rounded-full bg-rose-500 text-base text-white hover:bg-rose-600"
                >
                  {busy ? "Входим…" : "Присоединиться"}
                </Button>
              </form>
            </TabsContent>
          </Tabs>
          {error && (
            <p className="mt-4 text-center text-sm text-rose-600">{error}</p>
          )}
        </motion.div>
      </div>
    </main>
  );
}
