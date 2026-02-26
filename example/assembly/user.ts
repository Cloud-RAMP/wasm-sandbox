import { Context, debug } from "./sdk";
import { WSEvent } from "./protocol";

export function onMessage(event: WSEvent): void {
  const ctx = new Context();

  debug("User " + event.connectionId + " called onMessage");

  const res = ctx.fetch("helloo", "GET", "hello");
  debug("fetch response: " + res);
}

export function onJoin(event: WSEvent): void {
  debug("onJoin called!");
}

export function onLeave(event: WSEvent): void {
  debug("onLeave called!");
}

export function onError(event: WSEvent): void {
  debug("onError called!");
}