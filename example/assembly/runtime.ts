import { onMessage } from "./user";

// Internal function to be called by the WebAssembly
export function __onMessage(ptr: usize, len: usize): void {
  const msg = String.UTF8.decodeUnsafe(ptr, len);

  onMessage(msg);
}