import * as env from "./env";

export function debug(msg: string): void {
    env._debug(to_usize(msg), msg.length);
}

/**
 * The "Context" class provides an abstraction which the user can use
 * to call methods with external effects.
 * 
 * It has three main partitions that define related methods:
 * * store: contains methods for redis operations
 * * db: contains methods for persistent data operations
 * * room: contains methods for room operations
 * 
 * Example usage:
 * ```TypeScript
 * const ctx = new Context();
 * const usersResp = ctx.room.getAllUsers();
 * ...
 * ```
 */
export class Context {
  store: Store;
  db: DB;
  room: Room;

  constructor(){
    this.store = new Store();
    this.room = new Room();
    this.db = new DB();
  }
  
  /**
   * Logs a line to the application's logs
   * 
   * @param msg the message to be logged
   * @returns a status object indicating the success of the operation
   */
  log(msg: string): Status {
    const errPtr = env._log(to_usize(msg), msg.length);
    return get_status(errPtr);
  }

  /**
   * Fetches data from an external source
   * 
   * @param url url to be fetched from
   * @param method HTTP method (GET, POST, PUT, etc.)
   * @param body request body
   * @returns A result of the response body or failure
   */
  fetch(url: string, method: string, body: string): Result<string> {
    const valPtr = env._fetch(to_usize(url), url.length, to_usize(method), method.length, to_usize(body), body.length);
    return get_result(valPtr);
  }

  /**
   * Send a message to the given recipient from "SEVER" instead of your connection ID
   * 
   * @param recipient the user to send it to
   * @param message the message to send to the user
   * @returns a status representing the success of the operation
   */
  serverMessage(recipient: string, message: string): Status {
    const errPtr = env._sendMessage(to_usize(recipient), recipient.length, to_usize(message), message.length);
    return get_status(errPtr);
  }
}

/**
 * Defines methods that modify the current instance's redis
 */
class Store {
  /**
   * Set a key to the corresponding value
   * 
   * @param key the key to set
   * @param value the value to set it to
   * @returns A status representing the success of the operation
   */
  set(key: string, value: string): Status {
    const errPtr = env._set(to_usize(key), key.length, to_usize(value), value.length);
    return get_status(errPtr);
  }

  /**
   * Get the value of a key
   * 
   * @param key the key to retrieve
   * @returns A result containing the value or an error
   */
  get(key: string): Result<string> {    
    const valPtr = env._get(to_usize(key), key.length);
    return get_result(valPtr);
  }

  /**
   * Delete a key from the store
   * 
   * @param key the key to delete
   * @returns A status representing the success of the operation
   */
  del(key: string): Status {
    const valPtr = env._del(to_usize(key), key.length);
    return get_status(valPtr);
  }
}

class DB {
  /**
   * Set a key-value pair in the database
   * 
   * @param key the key to set
   * @param value the value to set it to
   * @returns A status representing the success of the operation
   */
  set(key: string, value: string): Status {
    const errPtr = env._dbSet(to_usize(key), key.length, to_usize(value), value.length);
    return get_status(errPtr);
  }

  /**
   * Get the value of a key from the database
   * 
   * @param key the key to retrieve
   * @returns A result containing the value or an error
   */
  get(key: string): Result<string> {    
    const valPtr = env._dbGet(to_usize(key), key.length);
    return get_result(valPtr);
  }

  /**
   * Delete a key from the database
   * 
   * @param key the key to delete
   * @returns A status representing the success of the operation
   */
  del(key: string): Status {
    const valPtr = env._dbDel(to_usize(key), key.length);
    return get_status(valPtr);
  }
}

class Room {
  /**
   * Broadcast a message to all users in the room
   * 
   * @param msg the message to broadcast
   * @returns A status representing the success of the operation
   */
  broadcast(msg: string): Status {
    const errPtr = env._broadcast(to_usize(msg), msg.length);
    return get_status(errPtr);
  }

  /**
   * Get a list of all users in the room
   * 
   * @returns A result containing an array of user IDs or an error
   */
  getUsers(): Result<string[]> {
    const ptr = env._getUsers();

    const buf = changetype<ArrayBuffer>(ptr);
    const users = decodeStringArray(buf);

    return users;
  }

  /**
   * Send a message to a specific user
   * 
   * @param recipient the ID of the recipient
   * @param message the message to send
   * @returns A status representing the success of the operation
   */
  sendMessage(recipient: string, message: string): Status {
    const errPtr = env._sendMessage(to_usize(recipient), recipient.length, to_usize(message), message.length);
    return get_status(errPtr);
  }

  /**
   * Close the connection for a specific user
   * 
   * @param target the ID of the user to disconnect
   * @returns A status representing the success of the operation
   */
  closeConnection(target: string): Status {
    const errPtr = env._closeConnection(to_usize(target), target.length);
    return get_status(errPtr);
  }
}


// AssemblyScript doesn't seem to allow for object interfaces so we need to use a class

/**
 * This class defines the information that is passed in wich each WS request.
 * 
 * Fields:
 * * connectionId: the unique identifier of the sending connection
 * * roomId: the unique identifier of the room in which the message was sent.
 *     * ! Can be empty if the connection is not a member of any room
 * * payload: the string of information that the user sent initially.
 *     * If you want to parse as JSON, consider using an AssemblyScript JSON library
 * * timestamp
 *     * the unixmilli timestamp of when the message was received 
 */
export class WSEvent {
  connectionId: string = "";
  roomId: string = "";
  payload: string = "";
  timestamp: number = 0;
}

/**
 * The "Result" class is used when a method has a return value,
 * but may also error. It is inspired by the similarly named type in Rust.
 * 
 * Example usage:
 * ```Typescript
 * const result: Result = someAsyncOperation();
 * if (result.isError()) {
 *     console.log("Error on async operation! " + result.error);
 * } else {
 *     consooe.log("Successful operation, data is: " + result.data);
 * }
 * ```
 */
export class Result<T> {
    data: T;
    error: string;

    constructor(data: T, error: string = "") {
        this.data = data;
        this.error = error;
    }

    isError(): bool {
        return this.error !== "";
    }
}

/**
 * The "Status" class is used when a method has no return value,
 * but we want to know the status of the operation.
 * 
 * Example usage:
 * ```Typescript
 * const status: Status = someAsyncOperation();
 * if (status.isError()) {
 *     console.log("Error on async operation! " + status.error);
 * }
 * ```
 */
export class Status {
    error: string;

    constructor(error: string = "") {
        this.error = error;
    }

    isError(): bool {
        return this.error !== "";
    }

    isOk(): bool {
        return this.error === "";
    }
}

function decodeStringArray(buf: ArrayBuffer): Result<string[]> {
    const data = Uint8Array.wrap(buf);
    let offset = 0;

    const indicator = data[0];
    if (indicator !== 43) { // + is 43 in ascii
        // Since the indicator is an error, we know the data was passed in as a regular string
        // Length is stored at addr - 4
        const addr = changetype<usize>(buf);
        const len = load<i32>(addr - 4);
        const errMsg = String.UTF16.decodeUnsafe(addr + 2, len - 2);

        return new Result([], errMsg);
    }

    offset += 2;

    // Read the number of strings (4 bytes, little endian)
    const count = (
        data[offset] |
        (data[offset + 1] << 8) |
        (data[offset + 2] << 16) |
        (data[offset + 3] << 24)
    ) >>> 0;
    offset += 4;

    const result = new Array<string>(count);

    let i: u8 = 0
    for (i = 0; i < count; i++) {
        // Read string length (4 bytes, little endian)
        const len =
        (data[offset] |
            (data[offset + 1] << 8) |
            (data[offset + 2] << 16) |
            (data[offset + 3] << 24)) >>> 0;
        offset += 4;

        // Read string bytes
        const strBytes = data.subarray(offset, offset + len);
        result[i] = String.UTF8.decodeUnsafe(changetype<usize>(strBytes.buffer) + strBytes.byteOffset, len);
        offset += len;
    }

    return new Result(result);
}

export function decodeWSEvent(buf: ArrayBuffer): WSEvent {
    const data = decodeStringArray(buf);
    if (data.isError()) {
        return new WSEvent();
    }
    
    const strArray = data.data;
    const ret: WSEvent = new WSEvent();
    ret.connectionId = strArray[0];
    ret.roomId = strArray[1];
    ret.timestamp = parseInt(strArray[2]);
    ret.payload = strArray[3];

    return ret;
}

function to_usize(str: string): usize {
    const ptr = String.UTF8.encode(str);
    return changetype<usize>(ptr);
}

function get_status(ptr: u32): Status {
    if (ptr == 0) {
        return new Status();
    }

    const len = load<i32>(ptr - 4);
    const val = String.UTF16.decodeUnsafe(ptr + 2, len - 2);
    return new Status(val);
}

function get_result(ptr: u32): Result<string> {
    // assembly script strings have the length stored 4 bytes before the string itself
    const len = load<i32>(ptr - 4);

    const indicator = load<u8>(ptr);
    if (indicator != 43) { // 43 is ascii '+'
        const errorMsg = String.UTF16.decodeUnsafe(ptr + 2, len - 2);
        
        return new Result("", errorMsg);
    }

    const val = String.UTF16.decodeUnsafe(ptr + 2, len - 2);
    return new Result(val);
}