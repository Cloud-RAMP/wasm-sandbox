import { Context } from "./sdk";
import { WSEvent } from "./protocol";

export function onMessage(event: WSEvent): void {
  const ctx = new Context();

  ctx.room.broadcast("My msg was: " + event.payload);
}
