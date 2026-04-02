export * from "./runtime";

  // const ctx = new Context();
  // const usersResp = ctx.room.getUsers();
  // if (usersResp.isError()) {
  //   debug("Error fetching users: " + usersResp.error);
  //   return;
  // }

  // const users = usersResp.data;
  // debug("Users in room:\n" + users.join(","))
  // if (users.length == 0) {
  //   return;
  // }

  // ctx.room.sendMessage(users[0], "Hello from other connection")


//   const ctx = new Context();

//   debug("User " + event.connectionId + " called onMessage");

//   const resp = ctx.fetch("http://google.com", "GET", "");
//   if (resp.isError()) {
//     abort("Fetch failed: " + resp.error);
//   }