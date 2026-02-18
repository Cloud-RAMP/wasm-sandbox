import { Context } from "./sdk";

export function onMessage(msg: string): void {
  const ctx = new Context();

  ctx.store.set("hello", "hello2");

  const str = ctx.store.get("test");
  ctx.room.broadcast("I got: " + str);
}
