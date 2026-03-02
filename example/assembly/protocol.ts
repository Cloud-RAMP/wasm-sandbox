// AssemblyScript doesn't seem to allow for object interfaces so we need to use a class

import { debug } from "./sdk";

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

export function decodeStringArray(buf: ArrayBuffer): Result<string[]> {
    const data = Uint8Array.wrap(buf);
    let offset = 0;

    const indicator = data[0];
    if (indicator !== 43) { // + is 43 in ascii
        // Since the indicator is an error, we know the data was passed in as a regular string
        // Length is stored at addr - 4
        const addr = changetype<usize>(buf);
        const len = load<i32>(addr - 4);
        const errMsg = String.UTF16.decodeUnsafe(addr + 2, len - 2);

        // figure out how to get the error text in here
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

export function to_usize(str: string): usize {
    const ptr = String.UTF8.encode(str);
    return changetype<usize>(ptr);
}

export function get_external_string(ptr: u32): Result<string> {
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