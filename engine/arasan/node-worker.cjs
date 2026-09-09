// Real engine harness for tests and the developer bot lab, never the backend.
const { parentPort } = require("node:worker_threads");
const { readFileSync } = require("node:fs");
const { resolve } = require("node:path");
const folder = resolve(__dirname, "../../web/public/arasan");
const createArasan = new Function("require", "__dirname", readFileSync(resolve(folder, "arasan.js"), "utf8") + "; return createArasan;")(require, folder);
const commands = [];
parentPort.on("message", command => commands.push(command));
createArasan({ commands, wasmBinary: readFileSync(resolve(folder, "arasan.wasm")), print: line => parentPort.postMessage(line), printErr: line => parentPort.postMessage(line) })
 .then(module => module.callMain(["-c", "1", "-H", "16M"]));
