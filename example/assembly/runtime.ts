import { decodeWSEvent } from "./protocol";
import { onMessage } from "./user";

// Internal function to be called by the WebAssembly
//
// Find a way to conditionall import this, in case the user did not define an onMessage function
export function __onMessage(ptr: usize, len: usize): void {
  const buf = changetype<ArrayBuffer>(ptr);
  const msg = decodeWSEvent(buf);
  // const msg = String.UTF8.decodeUnsafe(ptr, len);

  onMessage(msg[0]);
}