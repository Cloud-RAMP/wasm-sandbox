import { decodeStringArray, get_result, to_usize, get_status, Result, Status } from "./protocol";
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