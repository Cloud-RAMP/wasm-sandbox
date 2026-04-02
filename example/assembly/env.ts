//@ts-ignore
@external("env", "set")
export declare function _set(keyPtr: usize, keyLen: usize, valPtr: usize, valLen: usize): usize;

//@ts-ignore
@external("env", "get")
export declare function _get(keyPtr: usize, keyLen: usize): usize;

//@ts-ignore
@external("env", "del")
export declare function _del(keyPtr: usize, keyLen: usize): usize;

//@ts-ignore
@external("env", "dbSet")
export declare function _dbSet(keyPtr: usize, keyLen: usize, valPtr: usize, valLen: usize): usize;

//@ts-ignore
@external("env", "dbGet")
export declare function _dbGet(keyPtr: usize, keyLen: usize): usize;

//@ts-ignore
@external("env", "dbDel")
export declare function _dbDel(keyPtr: usize, keyLen: usize): usize;

//@ts-ignore
@external("env", "log")
export declare function _log(msgPtr: usize, msgLen: usize): usize;

//@ts-ignore
@external("env", "debug")
export declare function _debug(msgPtr: usize, msgLen: usize): usize;

//@ts-ignore
@external("env", "broadcast")
export declare function _broadcast(ptr: usize, len: usize): usize;

//@ts-ignore
@external("env", "getUsers")
export declare function _getUsers(): usize;

//@ts-ignore
@external("env", "sendMessage")
export declare function _sendMessage(userPtr: usize, userLen: usize, msgPtr: usize, msgLen: usize): usize;

//@ts-ignore
@external("env", "closeConnection")
export declare function _closeConnection(targetPtr: usize, targetLen: usize): usize;

//@ts-ignore
@external("env", "fetch")
export declare function _fetch(urlPtr: usize, urlLen: usize, methodPtr: usize, methodLen: usize, bodyPtr: usize, bodyLen: usize): usize;