import { Context } from "./sdk";

export function onMessage(msg: string): void {
  const ctx = new Context();

  ctx.room.broadcast("My msg was: " + msg)
}
