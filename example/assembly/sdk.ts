import { decodeGoArray, get_external_string, to_usize } from "./protocol";
import * as env from "./env";

export function debug(msg: string): void {
    env._debug(to_usize(msg), msg.length);
}

export class Context {
  store: Store
  room: Room

  constructor(){
    this.store = new Store()
    this.room = new Room()
  }
  
  log(msg: string): void {
    env._log(to_usize(msg), msg.length);
  }
}

class Store {
  set(key: string, value: string): void {
    env._set(to_usize(key), key.length, to_usize(value), value.length)
  }

  get(key: string): string {    
    const valPtr = env._get(to_usize(key), key.length);
    return get_external_string(valPtr);
  }
}

class Room {
  broadcast(msg: string): void {
    env._broadcast(to_usize(msg), msg.length);
  }

  getUsers(): string[] {
    const ptr = env._getUsers();

    const buf = changetype<ArrayBuffer>(ptr);
    const users = decodeGoArray(buf);

    return users
  }
}