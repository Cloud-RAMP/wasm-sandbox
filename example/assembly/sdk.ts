@external("env", "broadcast")
declare function _broadcast(ptr: usize, len: usize): void;

export class Context {
    broadcast(msg: string): void {
    const ptr = String.UTF8.encode(msg);
    _broadcast(changetype<usize>(ptr), msg.length);
  }
}