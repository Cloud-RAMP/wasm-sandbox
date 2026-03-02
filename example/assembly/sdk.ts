import { decodeStringArray, get_external_string, to_usize, Result } from "./protocol";
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

  fetch(url: string, method: string, body: string): Result<string> {
    const valPtr = env._fetch(to_usize(url), url.length, to_usize(method), method.length, to_usize(body), body.length);
    return get_external_string(valPtr);
  }
}

class Store {
  set(key: string, value: string): void {
    env._set(to_usize(key), key.length, to_usize(value), value.length)
  }

  get(key: string): Result<string> {    
    const valPtr = env._get(to_usize(key), key.length);
    return get_external_string(valPtr);
  }
}

class Room {
  broadcast(msg: string): void {
    env._broadcast(to_usize(msg), msg.length);
  }

  getUsers(): Result<string[]> {
    const ptr = env._getUsers();

    const buf = changetype<ArrayBuffer>(ptr);
    const users = decodeStringArray(buf);

    return users
  }

  sendMessage(recipient: string, message: string): void {
    env._sendMessage(to_usize(recipient), recipient.length, to_usize(message), message.length, )
  }
}