import { Context, debug } from "./sdk";
import { WSEvent } from "./protocol";

export function onMessage(event: WSEvent): void {
  const ctx = new Context();

  // debug("User " + event.connectionId + " called onMessage");

  const resp = ctx.fetch("http://google.com", "GET", "");
  if (resp.isError()) {
    abort("Fetch failed: " + resp.error);
  }
}

export function onJoin(event: WSEvent): void {
  debug("User " + event.connectionId + " called onJoin");
}

export function onLeave(event: WSEvent): void {
  debug("User " + event.connectionId + " called onLeave");
}

export function onError(event: WSEvent): void {
  debug("User " + event.connectionId + " called onError");
}