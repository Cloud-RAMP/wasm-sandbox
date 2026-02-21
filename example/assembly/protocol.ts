import { debug } from "./sdk";

// AssemblyScript doesn't seem to allow for interfaces so we need to use a class
export class WSEvent {
  connectionId: string = "";
  roomId: string = "";
  payload: string = "";
  timestamp: number = 0;
}

export function decodeWSEvent(buf: ArrayBuffer): WSEvent {

    const data = Uint8Array.wrap(buf);
    let offset = 0;

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
        debug(result[i]);
        offset += len;
    }
    
    const ret: WSEvent = new WSEvent();
    ret.connectionId = result[0];
    ret.roomId = result[1];
    ret.timestamp = parseInt(result[2]);
    ret.payload = result[3];

    return ret;
}