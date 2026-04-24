import { Context, debug } from "./sdk";
import { WSEvent } from "./protocol";

export function onMessage(event: WSEvent): void {
  debug("User " + event.connectionId + " called onMessage");

  const ctx = new Context();
  ctx.room.broadcast(event.payload);
}

export function onJoin(event: WSEvent): void {
  debug("User " + event.connectionId + " called onJoin");

  const ctx = new Context();
  const usersRes = ctx.room.getUsers();
  if (usersRes.isError()) {
    debug("Failed to get users: " + usersRes.error);
    return
  }

  for (let i = 0; i < usersRes.data.length; i++) {
    debug(usersRes.data[i]);
  }
}

export function onLeave(event: WSEvent): void {
  debug("User " + event.connectionId + " called onLeave");
}

export function onError(event: WSEvent): void {
  debug("User " + event.connectionId + " called onError");
}