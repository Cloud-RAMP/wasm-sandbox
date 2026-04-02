import { decodeStringArray, get_result, to_usize, get_status, Result, Status } from "./protocol";
import * as env from "./env";

export function debug(msg: string): void {
    env._debug(to_usize(msg), msg.length);
}

export class Context {
  store: Store;
  db: DB;
  room: Room;

  constructor(){
    this.store = new Store();
    this.room = new Room();
    this.db = new DB();
  }
  
  log(msg: string): Status {
    const errPtr = env._log(to_usize(msg), msg.length);
    return get_status(errPtr);
  }

  fetch(url: string, method: string, body: string): Result<string> {
    const valPtr = env._fetch(to_usize(url), url.length, to_usize(method), method.length, to_usize(body), body.length);
    return get_result(valPtr);
  }
}

class Store {
  set(key: string, value: string): Status {
    const errPtr = env._set(to_usize(key), key.length, to_usize(value), value.length);
    return get_status(errPtr);
  }

  get(key: string): Result<string> {    
    const valPtr = env._get(to_usize(key), key.length);
    return get_result(valPtr);
  }

  del(key: string): Status {
    const valPtr = env._del(to_usize(key), key.length);
    return get_status(valPtr);
  }
}

class DB {
  set(key: string, value: string): Status {
    const errPtr = env._dbSet(to_usize(key), key.length, to_usize(value), value.length);
    return get_status(errPtr);
  }

  get(key: string): Result<string> {    
    const valPtr = env._dbGet(to_usize(key), key.length);
    return get_result(valPtr);
  }

  del(key: string): Status {
    const valPtr = env._dbDel(to_usize(key), key.length);
    return get_status(valPtr);
  }
}

class Room {
  broadcast(msg: string): Status {
    const errPtr = env._broadcast(to_usize(msg), msg.length);
    return get_status(errPtr);
  }

  getUsers(): Result<string[]> {
    const ptr = env._getUsers();

    const buf = changetype<ArrayBuffer>(ptr);
    const users = decodeStringArray(buf);

    return users;
  }

  sendMessage(recipient: string, message: string): Status {
    const errPtr = env._sendMessage(to_usize(recipient), recipient.length, to_usize(message), message.length);
    return get_status(errPtr);
  }

  closeConnection(target: string): Status {
    const errPtr = env._closeConnection(to_usize(target), target.length);
    return get_status(errPtr);
  }
}