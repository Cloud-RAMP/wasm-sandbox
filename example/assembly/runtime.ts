import { decodeWSEvent } from "./protocol";
import { debug } from "./sdk";
import * as user from "./user";

// Internal function to be called by the WebAssembly
//
// Find a way to conditional import this, in case the user did not define an onMessage function
export function __onMessage(ptr: usize, len: usize): void {
  const buf = changetype<ArrayBuffer>(ptr);
  const event = decodeWSEvent(buf);
  // const msg = String.UTF8.decodeUnsafe(ptr, len);

  if (typeof user.onMessage === "function") {
    user.onMessage(event);
  }
}