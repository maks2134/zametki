import { getWsUrl } from "./api";
import type { WSEvent } from "./types";

type Handler = (event: WSEvent) => void;

export type ConnectionStatus = "connecting" | "online" | "offline";

export class RoomSocket {
  private ws: WebSocket | null = null;
  private token: string;
  private onEvent: Handler;
  private onStatus: (s: ConnectionStatus) => void;
  private closed = false;
  private retries = 0;
  private timer: ReturnType<typeof setTimeout> | null = null;

  constructor(
    token: string,
    onEvent: Handler,
    onStatus: (s: ConnectionStatus) => void,
  ) {
    this.token = token;
    this.onEvent = onEvent;
    this.onStatus = onStatus;
  }

  connect() {
    this.closed = false;
    this.onStatus("connecting");
    const ws = new WebSocket(getWsUrl(this.token));
    this.ws = ws;

    ws.onopen = () => {
      this.retries = 0;
      this.onStatus("online");
    };

    ws.onmessage = (msg) => {
      try {
        const event = JSON.parse(msg.data as string) as WSEvent;
        this.onEvent(event);
      } catch {
        // ignore malformed
      }
    };

    ws.onclose = () => {
      this.onStatus("offline");
      this.ws = null;
      if (!this.closed) this.scheduleReconnect();
    };

    ws.onerror = () => {
      ws.close();
    };
  }

  private scheduleReconnect() {
    const delay = Math.min(1000 * 2 ** this.retries, 15000);
    this.retries += 1;
    this.timer = setTimeout(() => this.connect(), delay);
  }

  disconnect() {
    this.closed = true;
    if (this.timer) clearTimeout(this.timer);
    this.ws?.close();
    this.ws = null;
    this.onStatus("offline");
  }
}
