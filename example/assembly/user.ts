import { Context, debug } from "./sdk";
import { WSEvent } from "./protocol";

export function onMessage(event: WSEvent): void {
  debug("User " + event.connectionId + " called onMessage");
}

export function onJoin(event: WSEvent): void {
  debug("User " + event.connectionId + " called onJoin");

  const ctx = new Context();
  const setRes = ctx.db.set("userId", event.connectionId);
  if (setRes.isError()) {
    debug("Error setting userId: " + setRes.error);
    return;
  }

  const getRes = ctx.db.get("userId");
  if (getRes.isError()) {
    debug("Error getting userId: " + getRes.error);
    return;
  }

  debug("User ID fetched from database: " + getRes.data);
}

export function onLeave(event: WSEvent): void {
  debug("User " + event.connectionId + " called onLeave");

  const ctx = new Context();
  const delRes = ctx.db.del("userId");
  if (delRes.isError()) {
    debug("Error setting userId: " + delRes.error);
    return;
  }
}

export function onError(event: WSEvent): void {
  debug("User " + event.connectionId + " called onError");
}