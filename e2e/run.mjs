// End to end tests: a real server binary, real Chrome, one human phone plus bots.
// Usage: node run.mjs [path-to-binary]. Set CHROME to override the browser path.
import { spawn } from "node:child_process";
import { chromium } from "playwright-core";
import net from "node:net";

const BIN = process.argv[2] || "../couchcade";
const CHROME = process.env.CHROME || "/usr/bin/google-chrome";

const freePort = () => new Promise((res) => { const s = net.createServer().listen(0, () => { const p = s.address().port; s.close(() => res(p)); }); });

async function startServer() {
  const port = await freePort();
  const proc = spawn(BIN, ["serve", "-addr", `127.0.0.1:${port}`, "-public-url", `http://127.0.0.1:${port}`, "-time-scale", "8"], { stdio: "ignore" });
  const base = `http://127.0.0.1:${port}`;
  for (let i = 0; i < 50; i++) {
    try { if ((await fetch(base + "/healthz")).ok) return { base, proc }; } catch { /* not up yet */ }
    await new Promise((r) => setTimeout(r, 100));
  }
  throw new Error("server did not start");
}

const results = [];
async function test(name, fn) {
  const t0 = Date.now();
  try { await fn(); results.push([name, true, Date.now() - t0]); console.log(`ok   ${name} (${Date.now() - t0}ms)`); }
  catch (e) { results.push([name, false, 0]); console.log(`FAIL ${name}\n     ${e.message.split("\n")[0]}`); }
}
const assert = (c, m) => { if (!c) throw new Error(m); };

const { base, proc } = await startServer();
const browser = await chromium.launch({ executablePath: CHROME, args: ["--no-sandbox"] });

async function room(opts = {}) {
  const errors = [];
  const watch = (p, who) => { p.on("pageerror", (e) => errors.push(`${who}: ${e.message}`)); };
  const tvCtx = await browser.newContext({ viewport: { width: 1280, height: 720 }, ...opts.tv });
  const tv = await tvCtx.newPage();
  watch(tv, "tv");
  await tv.goto(base + "/");
  await tv.waitForSelector(".code span");
  const code = (await tv.locator(".code span").allTextContents()).join("");
  const join = async (name, extra = {}) => {
    const ctx = await browser.newContext({ viewport: { width: 390, height: 780 }, hasTouch: true, isMobile: true, ...extra });
    const ph = await ctx.newPage();
    watch(ph, name);
    await ph.goto(`${base}/play?c=${code}`);
    await ph.fill("input:not(.code)", name);
    await ph.click(".btn.primary");
    await ph.waitForSelector(".me-card");
    return ph;
  };
  const close = async () => { for (const c of browser.contexts()) await c.close(); };
  return { tv, code, join, errors, close };
}

// A generic human: looks at the phone and does the obvious thing.
async function human(ph, stop) {
  let words = 0;
  while (!stop.done) {
    try {
      if (await ph.locator(".pad").count()) {
        const send = ph.locator(".btn.primary.wide");
        if (await send.isDisabled()) {
          const b = await ph.locator(".pad").boundingBox();
          await ph.mouse.move(b.x + 40, b.y + 40);
          await ph.mouse.down();
          await ph.mouse.move(b.x + b.width - 40, b.y + 80, { steps: 6 });
          await ph.mouse.move(b.x + 60, b.y + b.height - 40, { steps: 6 });
          await ph.mouse.up();
        }
        if (!(await send.isDisabled()) && !(await ph.locator(".hint.ok").count())) await send.click();
      } else if (await ph.locator(".field:not(.code)").count()) {
        if (!(await ph.locator(".hint.ok").count())) {
          await ph.fill(".field", ["pickle", "storm", "ember", "tiger"][words++ % 4] + words);
          await ph.click(".btn.primary.wide");
        }
      } else if (await ph.locator(".zapbtn.go").count()) {
        await ph.locator(".zapbtn.go").dispatchEvent("pointerdown");
      } else {
        for (const sel of [".tile:not(:disabled)", ".opt:not(:disabled):not(.picked)", ".vote:not(.picked)", ".card.art:not(:disabled):not(.picked)"]) {
          const el = ph.locator(sel).first();
          if (await el.count()) { await el.click({ timeout: 800 }); break; }
        }
      }
    } catch { /* the page changed under us, try again */ }
    await new Promise((r) => setTimeout(r, 250));
  }
}

const GAMES = ["squiggle", "fibber", "trivia", "zap", "echo", "mole"];
for (const [idx, game] of GAMES.entries()) {
  await test(`full ${game} game with a human and bots`, async () => {
    const r = await room();
    const ph = await r.join("Arthur");
    await r.tv.click("text=Add 3 bots");
    await r.tv.selectOption("select >> nth=-1", "2");
    await r.tv.locator(".gamecard").nth(idx).click();
    await r.tv.waitForSelector(`.g-${game}`, { timeout: 5000 });
    const stop = { done: false };
    const bot = human(ph, stop);
    try { await r.tv.waitForSelector(".final", { timeout: 90000 }); } finally { stop.done = true; await bot; }
    await r.tv.waitForFunction(() => [...document.querySelectorAll(".podium b, .rest b")].some((b) => b.textContent === "Arthur"), null, { timeout: 3000 })
      .catch(() => { throw new Error("human missing from final scores"); });
    const pts = await ph.locator(".mega").textContent();
    assert(/^#\d$/.test(pts), "phone shows no rank: " + pts);
    assert(!r.errors.length, r.errors.join("; "));
    await r.close();
  });
}

await test("party standings carry across games and can be reset", async () => {
  const r = await room();
  const ph = await r.join("Arthur");
  await r.tv.click("text=Add 3 bots");
  await r.tv.selectOption("select >> nth=-1", "2");
  const play = async () => {
    await r.tv.locator(".gamecard").nth(3).click();
    await r.tv.waitForSelector(".final", { timeout: 90000 });
  };
  await play();
  assert((await r.tv.locator(".final .total").count()) === 0, "total should wait for a second game");
  await r.tv.click("text=Back to lobby");
  await r.tv.waitForSelector(".standings");
  assert((await r.tv.locator(".lobby .chip .sc").count()) === 4, "every chip should show its total");
  await play();
  assert((await r.tv.locator(".final .total").count()) > 0, "final screen shows no party total");
  assert(/#\d/.test(await ph.locator(".mega").textContent()), "phone shows no rank");
  assert(await ph.locator("text=Party total").count(), "phone shows no party total");
  await r.tv.click("text=Play again");
  await r.tv.waitForSelector(".g-zap", { timeout: 5000 });
  await r.tv.waitForSelector(".final", { timeout: 90000 });
  await r.tv.click("text=Back to lobby");
  await r.tv.click("text=Reset standings");
  await r.tv.waitForSelector(".standings", { state: "detached" });
  assert(!r.errors.length, r.errors.join("; "));
  await r.close();
});

await test("phones keep the screen awake, with a video fallback on plain http", async () => {
  const r = await room();
  const ph = await r.join("Arthur", {});
  await ph.waitForTimeout(500);
  assert((await ph.locator("video").count()) === 0 || (await ph.evaluate(() => "wakeLock" in navigator)), "no wake lock and no fallback");
  const ctx = await browser.newContext({ viewport: { width: 390, height: 780 }, hasTouch: true, isMobile: true });
  await ctx.addInitScript(() => { Object.defineProperty(navigator, "wakeLock", { value: undefined }); delete Navigator.prototype.wakeLock; });
  const p2 = await ctx.newPage();
  await p2.goto(`${base}/play?c=${r.code}`);
  await p2.fill("input:not(.code)", "Mina");
  await p2.click(".btn.primary");
  await p2.waitForSelector(".me-card");
  await p2.waitForSelector("video", { state: "attached" });
  await p2.waitForFunction(() => { const v = document.querySelector("video"); return v && !v.paused; }, null, { timeout: 5000 });
  assert(!r.errors.length, r.errors.join("; "));
  await r.close();
});

await test("reconnect keeps the same player", async () => {
  const r = await room();
  const ph = await r.join("Mina");
  await ph.reload();
  await ph.waitForSelector(".me-card");
  assert((await r.tv.locator(".chip").count()) === 1, "reload created a second player");
  await r.close();
});

await test("wrong code and empty name are refused", async () => {
  const r = await room();
  const ctx = await browser.newContext({ viewport: { width: 390, height: 780 } });
  const ph = await ctx.newPage();
  await ph.goto(`${base}/play?c=ZZZZ`);
  await ph.fill("input:not(.code)", "Lost");
  await ph.click(".btn.primary");
  await ph.waitForSelector("[role=alert]");
  await r.close();
});

await test("host phone can kick and the kicked phone is told", async () => {
  const r = await room();
  const a = await r.join("Host");
  const b = await r.join("Guest");
  await a.click("text=Manage players");
  await a.click("text=Kick");
  await b.waitForSelector("[role=alert], .joinform", { timeout: 4000 });
  await r.tv.waitForFunction(() => document.querySelectorAll(".chip").length === 1, null, { timeout: 4000 }).catch(() => { throw new Error("guest still seated"); });
  await r.close();
});

await test("players beyond eight join the audience", async () => {
  const r = await room();
  for (let i = 0; i < 9; i++) await r.join("P" + i);
  assert((await r.tv.locator(".chip").count()) === 8, "expected 8 seats");
  await r.tv.waitForSelector("text=1 watching");
  await r.close();
});

await test("light and dark themes both render", async () => {
  const r = await room();
  const t0 = await r.tv.evaluate(() => document.documentElement.dataset.theme);
  await r.tv.click("[aria-label=Theme]");
  const t1 = await r.tv.evaluate(() => document.documentElement.dataset.theme);
  assert(t0 !== t1, "theme did not change");
  await r.close();
});

await test("phone layout has no sideways scroll in every game", async () => {
  const r = await room();
  const ph = await r.join("Narrow", { viewport: { width: 320, height: 640 } });
  await r.tv.click("text=Add 3 bots");
  for (let i = 0; i < GAMES.length; i++) {
    await r.tv.locator(".gamecard").nth(i).click();
    await ph.waitForSelector(`.g-${GAMES[i]}`);
    await ph.waitForTimeout(700);
    const over = await ph.evaluate(() => document.documentElement.scrollWidth - innerWidth);
    assert(over <= 0, `${GAMES[i]} overflows by ${over}px`);
    await r.tv.click("text=End game");
    await ph.waitForSelector(".me-card");
  }
  await r.close();
});

await test("every language loads and translates the lobby", async () => {
  const r = await room();
  for (const l of ["zh-TW", "zh-CN", "ja", "ko", "es", "fr", "de", "pt-BR"]) {
    await r.tv.selectOption("select >> nth=0", l);
    await r.tv.waitForFunction((l) => document.documentElement.lang === l, l);
    const label = await r.tv.locator(".side h2").textContent();
    assert(!/^Players/.test(label), `${l} lobby still English: ${label}`);
  }
  await r.close();
});

await browser.close();
proc.kill();
const failed = results.filter((r) => !r[1]).length;
console.log(`\n${results.length - failed}/${results.length} passed`);
process.exit(failed ? 1 : 0);
