import { Context, debug } from "./sdk";
import { WSEvent } from "./protocol";

export function onMessage(event: WSEvent): void {
  const ctx = new Context();
  ctx.room.broadcast(event.payload);
}

export function onJoin(event: WSEvent): void {
  const ctx = new Context();
  ctx.room.broadcast("New user " + event.connectionId + " joined!");
}

export function onLeave(event: WSEvent): void {
  const ctx = new Context();
  ctx.room.broadcast("User " + event.connectionId + " left");
}

export function onError(event: WSEvent): void {
}