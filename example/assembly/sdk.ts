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
declare function _log(msgPtr: usize, msgLen: usize): usize;

export class Context {
  store: Store
  room: Room

  constructor(){
    this.store = new Store()
    this.room = new Room()
  }

  log(msg: string): void {
    const msgPtr = String.UTF8.encode(msg);

    _log(changetype<usize>(msgPtr), msg.length);
  }
}

class Store {
  set(key: string, value: string): void {
    const keyPtr = String.UTF8.encode(key);
    const valPtr = String.UTF8.encode(value);

    _set(changetype<usize>(keyPtr), key.length, changetype<usize>(valPtr), value.length)
  }

  get(key: string): string {
    const keyPtr = String.UTF8.encode(key);
    
    const valPtr = _get(changetype<usize>(keyPtr), key.length);
    
    // assembly script strings have the length stored 4 bytes before the string itself
    const valLen = load<i32>(valPtr - 4);

    const val = String.UTF16.decodeUnsafe(valPtr, valLen);
    return val
  }
}

class Room {
  broadcast(msg: string): void {
    const ptr = String.UTF8.encode(msg);
    _broadcast(changetype<usize>(ptr), msg.length);
  }
}