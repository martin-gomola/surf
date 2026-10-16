import {readFile} from "node:fs/promises";
import https from "node:https";

const descriptor = JSON.parse(await readFile("/surf-home/daemon.json", "utf8"));
const backendStats = () => new Promise((resolve, reject) => {
  const url = new URL("/api/v1/health?stats=1", descriptor.controlURL);
  const request = https.get(url, {
    rejectUnauthorized: false,
    headers: {"X-Surf-Admin": descriptor.adminToken},
  }, response => {
    let body = "";
    response.setEncoding("utf8");
    response.on("data", chunk => { body += chunk; });
    response.on("end", () => {
      if (response.statusCode !== 200) reject(new Error(`Surf stats HTTP ${response.statusCode}: ${body}`));
      else {
        try { resolve(JSON.parse(body)); }
        catch (error) { reject(error); }
      }
    });
  });
  request.on("error", reject);
});

const durationSeconds = Math.max(5, Number(process.argv[2] || 60));
const port = (await readFile("/profile/DevToolsActivePort", "utf8")).split("\n")[0];
const targets = await fetch(`http://127.0.0.1:${port}/json/list`).then((response) => response.json());
const page = targets.find((target) => target.type === "page" && /^https?:/.test(target.url));
if (!page) throw new Error("Surf has no web page target to benchmark");

const socket = new WebSocket(page.webSocketDebuggerUrl);
await new Promise((resolve, reject) => {
  socket.addEventListener("open", resolve, {once: true});
  socket.addEventListener("error", reject, {once: true});
});

let sequence = 0;
const pending = new Map();
socket.addEventListener("message", (event) => {
  const message = JSON.parse(event.data);
  if (!message.id || !pending.has(message.id)) return;
  const {resolve, reject} = pending.get(message.id);
  pending.delete(message.id);
  if (message.error) reject(new Error(message.error.message));
  else resolve(message.result);
});
const call = (method, params = {}) => new Promise((resolve, reject) => {
  const id = ++sequence;
  pending.set(id, {resolve, reject});
  socket.send(JSON.stringify({id, method, params}));
});

const started = performance.now();
const before = await backendStats();
let direction = 1;
let nextTurn = started + 3000;
let events = 0;
while (performance.now() - started < durationSeconds * 1000) {
  const now = performance.now();
  if (now >= nextTurn) {
    direction *= -1;
    nextTurn = now + 3000;
  }
  await call("Input.dispatchMouseEvent", {
    type: "mouseWheel",
    x: 384,
    y: 487,
    deltaX: 0,
    deltaY: 120 * direction,
  });
  events++;
  await new Promise((resolve) => setTimeout(resolve, 50));
}

socket.close();
const after = await backendStats();
console.log(JSON.stringify({
  duration_seconds: durationSeconds,
  scroll_events: events,
  page: new URL(page.url).origin,
  server_before: before,
  server_after: after,
}));
