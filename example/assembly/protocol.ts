export function decodeWSEvent(buf: ArrayBuffer): string[] {

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
        result[i] = String.UTF8.decode(strBytes.buffer, true);
        offset += len;
    }

    return result;
}