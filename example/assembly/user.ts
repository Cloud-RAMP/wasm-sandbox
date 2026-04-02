import { Context, debug } from "./sdk";
import { WSEvent } from "./protocol";

export function onMessage(event: WSEvent): void {
  debug("User " + event.connectionId + " called onMessage");
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