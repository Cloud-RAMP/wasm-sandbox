import { Context } from "./sdk";

export function onMessage(msg: string): void {
  const ctx = new Context();

  const str = ctx.store.get("test");

  ctx.room.broadcast("I got this message: " + str)
}
