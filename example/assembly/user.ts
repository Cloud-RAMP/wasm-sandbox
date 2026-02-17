import { Context } from "./sdk";

export function onMessage(msg: string): void {
  const ctx = new Context();

  ctx.broadcast("I got your message: " + msg);
}
