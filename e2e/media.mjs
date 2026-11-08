// Records the hosted demo while a human plays one Squiggle round, then turns the video into assets/demo.gif.
// Usage: node media.mjs <site-dir> <out-dir>
import http from "node:http";
import fs from "node:fs";
import path from "node:path";
import { execFileSync } from "node:child_process";
import { chromium } from "playwright-core";

const dir = path.resolve(process.argv[2]);
const out = path.resolve(process.argv[3]);
fs.mkdirSync(out, { recursive: true });
const types = { ".html": "text/html", ".js": "text/javascript", ".css": "text/css", ".json": "application/json", ".svg": "image/svg+xml", ".wasm": "application/wasm", ".png": "image/png" };
const srv = http.createServer((req, res) => {
  let p = path.join(dir, decodeURIComponent(req.url.split("?")[0]));
  if (p.endsWith("/")) p += "index.html";
  if (!p.startsWith(dir) || !fs.existsSync(p)) { res.writeHead(404).end(); return; }
  res.writeHead(200, { "content-type": types[path.extname(p)] || "application/octet-stream" }).end(fs.readFileSync(p));
}).listen(0);
const base = `http://127.0.0.1:${srv.address().port}/`;

const browser = await chromium.launch({ executablePath: process.env.CHROME || "/usr/bin/google-chrome", args: ["--no-sandbox"] });
const ctx = await browser.newContext({ viewport: { width: 1280, height: 700 }, colorScheme: "dark" });
const page = await ctx.newPage();
// No video codec needed: grab frames in a loop and let ffmpeg stitch them at the measured rate.
const frames = path.join(out, "frames");
fs.rmSync(frames, { recursive: true, force: true });
fs.mkdirSync(frames, { recursive: true });
let n = 0, rec = true;
const t0 = Date.now();
const grabber = (async () => {
  while (rec) {
    try { await page.screenshot({ path: path.join(frames, String(n++).padStart(5, "0") + ".jpg"), type: "jpeg", quality: 80 }); } catch { /* page busy */ }
  }
})();
await page.goto(base);
const tv = page.frameLocator("iframe.tv");
const phone = page.frameLocator("iframe.phone").first();
await tv.locator(".code span").first().waitFor({ timeout: 20000 });
await phone.locator(".me-card").waitFor();
await page.waitForTimeout(1200);
await tv.locator("text=Add 3 bots").click();
await page.waitForTimeout(1500);
await tv.locator(".gamecard").nth(0).click();
await phone.locator(".pad").waitFor();
await page.waitForTimeout(800);
const box = await phone.locator(".pad").boundingBox();
const cx = box.x + box.width / 2, cy = box.y + box.height / 2;
async function stroke(pts, color) {
  if (color !== undefined) await phone.locator(".sw").nth(color).click();
  await page.mouse.move(cx + pts[0][0], cy + pts[0][1]);
  await page.mouse.down();
  for (const [x, y] of pts.slice(1)) await page.mouse.move(cx + x, cy + y, { steps: 5 });
  await page.mouse.up();
}
// a toaster with a ghost face
await stroke([[-70, -20], [70, -20], [70, 50], [-70, 50], [-70, -20]], 0);
await stroke([[-50, -20], [-50, -50], [-20, -50], [-20, -20]], 1);
await stroke([[20, -20], [20, -50], [50, -50], [50, -20]], 3);
await stroke([[-30, 10], [-30, 20]], 5);
await stroke([[30, 10], [30, 20]], 5);
await stroke([[-20, 35], [0, 42], [20, 35]], 8);
await page.waitForTimeout(500);
await phone.locator(".btn.primary.wide").click();
await tv.locator(".gallery").waitFor({ timeout: 40000 });
await page.waitForTimeout(2500);
await phone.locator(".card.art:not(:disabled)").first().click();
await tv.locator(".card.art .by").first().waitFor({ timeout: 40000 });
await page.waitForTimeout(3500);
await page.screenshot({ path: path.join(out, "reveal.png") });
rec = false;
await grabber;
const fps = n / ((Date.now() - t0) / 1000);
await ctx.close();
await browser.close();
srv.close();
const pal = path.join(out, "pal.png");
const vf = "scale=860:-1:flags=lanczos";
const inp = ["-framerate", fps.toFixed(2), "-i", path.join(frames, "%05d.jpg")];
execFileSync("ffmpeg", ["-y", "-v", "error", ...inp, "-vf", `${vf},palettegen=max_colors=96:stats_mode=diff`, pal]);
execFileSync("ffmpeg", ["-y", "-v", "error", ...inp, "-i", pal, "-lavfi", `${vf}[x];[x][1:v]paletteuse=dither=bayer:bayer_scale=4`, "-r", "10", path.join(out, "demo.gif")]);
console.log("done", n, "frames", fps.toFixed(1), "fps", fs.statSync(path.join(out, "demo.gif")).size);
