//@ts-ignore
@external("env", "set")
export declare function _set(keyPtr: usize, keyLen: usize, valPtr: usize, valLen: usize): void;

//@ts-ignore
@external("env", "get")
export declare function _get(keyPtr: usize, keyLen: usize): usize;

//@ts-ignore
@external("env", "log")
export declare function _log(msgPtr: usize, msgLen: usize): void;

//@ts-ignore
@external("env", "debug")
export declare function _debug(msgPtr: usize, msgLen: usize): void;

//@ts-ignore
@external("env", "broadcast")
export declare function _broadcast(ptr: usize, len: usize): void;

//@ts-ignore
@external("env", "getUsers")
export declare function _getUsers(): usize;

//@ts-ignore
@external("env", "sendMessage")
export declare function _sendMessage(userPtr: usize, userLen: usize, msgPtr: usize, msgLen: usize): void;