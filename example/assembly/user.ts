import { Context, debug } from "./sdk";
import { WSEvent } from "./protocol";

export function onMessage(event: WSEvent): void {
  const ctx = new Context();

  debug("User " + event.connectionId + " called onMessage");

  // let res = ctx.log("hello?");
  // if (res.isError()) {
  //   debug("log error: " + res.error);
  // } else {
  //   debug("log successful");
  // }
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