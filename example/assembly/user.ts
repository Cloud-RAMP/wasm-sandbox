import { Context, debug } from "./sdk";
import { WSEvent } from "./protocol";

export function onMessage(event: WSEvent): void {
  const ctx = new Context();

  ctx.room.broadcast("My msg was: " + event.payload);
}

export function onJoin(event: WSEvent): void {
  debug("onJoin called!");
}

export function onLeave(event: WSEvent): void {
  debug("onLeave called!");
}