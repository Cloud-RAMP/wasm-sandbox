export interface WSEvent {
    connectionId: string,
    roomId: string,
    timestamp: number,
    payload: string,
}
