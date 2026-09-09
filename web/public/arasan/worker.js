/* Arasan transport only. This worker never receives player credentials. */
importScripts("arasan.js");
const commands = [];
self.onmessage = event => {
  if (typeof event.data === "string" && event.data.length < 20000 && !/[\r\n]/.test(event.data)) commands.push(event.data);
};
createArasan({
  commands,
  print: line => self.postMessage(line),
  printErr: () => {},
  onAbort: () => self.postMessage("engine-error"),
}).then(module => module.callMain(["-c", "1", "-H", "16M"]))
  .catch(() => self.postMessage("engine-error"));
