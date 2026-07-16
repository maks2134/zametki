"use client";

import { use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Heart } from "lucide-react";
import { joinRoom } from "@/lib/api";
import { getToken, setSession } from "@/lib/auth";
import { MEMBER_COLORS } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function JoinPage({
  params,
}: {
  params: Promise<{ code: string }>;
}) {
  const { code: inviteCode } = use(params);
  const router = useRouter();
  const [name, setName] = useState("");
  const [color, setColor] = useState(MEMBER_COLORS[1]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (getToken()) router.replace("/room");
  }, [router]);

  async function onJoin(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const res = await joinRoom({
        code: inviteCode.trim().toUpperCase(),
        name: name.trim(),
        color,
      });
      const me = res.room.members[res.room.members.length - 1];
      setSession(res.token, res.room, me);
      router.push("/room");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось войти");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="flex min-h-dvh items-center justify-center bg-[radial-gradient(ellipse_at_top,_#ffd6e0,_transparent_55%),#fff7f8] px-4 py-8 pt-[max(2rem,env(safe-area-inset-top))] pb-[max(2rem,env(safe-area-inset-bottom))]">
      <form
        onSubmit={(e) => void onJoin(e)}
        className="w-full max-w-md space-y-5 rounded-[1.75rem] border border-white/80 bg-white/85 p-6 shadow-xl backdrop-blur sm:rounded-[2rem] sm:p-8"
      >
        <div className="text-center">
          <Heart className="mx-auto h-8 w-8 fill-rose-400 text-rose-400" />
          <h1 className="mt-3 font-[family-name:var(--font-display)] text-2xl text-rose-950 sm:text-3xl">
            Вас пригласили
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Код:{" "}
            <span className="font-mono tracking-widest">{inviteCode}</span>
          </p>
        </div>
        <div className="space-y-2">
          <Label htmlFor="name">Ваше имя</Label>
          <Input
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="h-12 rounded-xl text-base"
            autoComplete="nickname"
            required
          />
        </div>
        <div className="space-y-2">
          <Label>Цвет</Label>
          <div className="flex flex-wrap gap-3">
            {MEMBER_COLORS.map((c) => (
              <button
                key={c}
                type="button"
                aria-label={c}
                onClick={() => setColor(c)}
                className="h-11 w-11 rounded-full border-2 active:scale-95"
                style={{
                  backgroundColor: c,
                  borderColor: color === c ? "#1f0a12" : "transparent",
                }}
              />
            ))}
          </div>
        </div>
        {error && <p className="text-sm text-rose-600">{error}</p>}
        <Button
          type="submit"
          disabled={busy || !name.trim()}
          className="h-12 w-full rounded-full bg-rose-500 text-base text-white hover:bg-rose-600"
        >
          {busy ? "Входим…" : "Присоединиться"}
        </Button>
      </form>
    </main>
  );
}
