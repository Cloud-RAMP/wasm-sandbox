import { get_external_string, to_usize } from "./protocol";

//@ts-ignore
@external("env", "broadcast")
declare function _broadcast(ptr: usize, len: usize): void;

//@ts-ignore
@external("env", "set")
declare function _set(keyPtr: usize, keyLen: usize, valPtr: usize, valLen: usize): void;

//@ts-ignore
@external("env", "get")
declare function _get(keyPtr: usize, keyLen: usize): usize;

//@ts-ignore
@external("env", "log")
declare function _log(msgPtr: usize, msgLen: usize): void;

//@ts-ignore
@external("env", "debug")
declare function _debug(msgPtr: usize, msgLen: usize): void;

export function debug(msg: string): void {
    _debug(to_usize(msg), msg.length);
}

export class Context {
  store: Store
  room: Room

  constructor(){
    this.store = new Store()
    this.room = new Room()
  }
  
  log(msg: string): void {
    _log(to_usize(msg), msg.length);
  }
}

class Store {
  set(key: string, value: string): void {
    _set(to_usize(key), key.length, to_usize(value), value.length)
  }

  get(key: string): string {    
    const valPtr = _get(to_usize(key), key.length);
    return get_external_string(valPtr);
  }
}

class Room {
  broadcast(msg: string): void {
    _broadcast(to_usize(msg), msg.length);
  }
}