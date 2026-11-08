// Checks the hosted demo: the engine runs as WebAssembly inside the page, no server involved.
// Usage: node demo.mjs <site-dir>
import http from "node:http";
import fs from "node:fs";
import path from "node:path";
import { chromium } from "playwright-core";

const assert = (c, m) => { if (!c) throw new Error(m); };
const dir = path.resolve(process.argv[2] || "../_site");
const types = { ".html": "text/html", ".js": "text/javascript", ".css": "text/css", ".json": "application/json", ".svg": "image/svg+xml", ".wasm": "application/wasm", ".png": "image/png" };
const srv = http.createServer((req, res) => {
  let p = path.join(dir, decodeURIComponent(req.url.split("?")[0]));
  if (p.endsWith("/")) p += "index.html";
  if (!p.startsWith(dir) || !fs.existsSync(p)) { res.writeHead(404).end(); return; }
  res.writeHead(200, { "content-type": types[path.extname(p)] || "application/octet-stream" }).end(fs.readFileSync(p));
}).listen(0);
const base = `http://127.0.0.1:${srv.address().port}/`;

const browser = await chromium.launch({ executablePath: process.env.CHROME || "/usr/bin/google-chrome", args: ["--no-sandbox"] });
const page = await browser.newPage({ viewport: { width: 1500, height: 800 } });
const errors = [];
page.on("pageerror", (e) => errors.push(e.message));
let ok = false;
try {
  await page.goto(base);
  const tv = page.frameLocator("iframe.tv");
  await tv.locator(".code span").first().waitFor({ timeout: 20000 });
  const phone = page.frameLocator("iframe.phone").first();
  await phone.locator(".me-card").waitFor({ timeout: 10000 });
  // The demo seats three bots by itself, so the games are ready to start.
  await tv.locator(".chip").nth(3).waitFor();
  assert(await tv.locator(".gamecard").nth(3).isEnabled(), "game cards are still locked");
  await tv.locator(".gamecard").nth(3).click();
  await tv.locator(".g-zap").waitFor({ timeout: 5000 });
  await phone.locator(".zapbtn").waitFor();
  await tv.locator(".final").waitFor({ timeout: 90000 });
  if (errors.length) throw new Error(errors.join("; "));
  ok = true;
  console.log("ok   hosted demo plays a full game in WebAssembly");
} catch (e) {
  console.log("FAIL hosted demo: " + e.message.split("\n")[0]);
} finally {
  await browser.close();
  srv.close();
}
process.exit(ok ? 0 : 1);
