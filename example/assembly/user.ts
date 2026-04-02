import { Context, debug } from "./sdk";
import { WSEvent } from "./protocol";

export function onMessage(event: WSEvent): void {
  debug("User " + event.connectionId + " called onMessage");
}

export function onJoin(event: WSEvent): void {
  debug("User " + event.connectionId + " called onJoin");

  const ctx = new Context();
  const usersResp = ctx.room.getUsers();
  if (usersResp.isError()) {
    debug("Failed to fetch users: " + usersResp.error);
    return;
  }

  // loop through all users and close connections that aren't ours
  const users = usersResp.data;
  for (let i = 0; i < users.length; i++) {
    const user = users[i];
    if (user != event.connectionId) {
      ctx.room.closeConnection(user);
    }
  }
}

export function onLeave(event: WSEvent): void {
  debug("User " + event.connectionId + " called onLeave");
}

export function onError(event: WSEvent): void {
  debug("User " + event.connectionId + " called onError");
}