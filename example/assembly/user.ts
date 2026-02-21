import { Context, debug } from "./sdk";
import { WSEvent } from "./protocol";

export function onMessage(event: WSEvent): void {
  const ctx = new Context();

  const users = ctx.room.getUsers();
  for (let i = 0; i < users.length; i++) {
    debug(users[i])
  }
}

export function onJoin(event: WSEvent): void {
  debug("onJoin called!");
}

export function onLeave(event: WSEvent): void {
  debug("onLeave called!");
}