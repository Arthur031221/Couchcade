// Drawing: a pointer driven pad for phones and an SVG renderer for finished drawings.
import { h } from "./dom.js";

export const INK = ["#1b1b2f", "#e63946", "#f77f00", "#fcbf49", "#7ac74f", "#2a9d8f", "#4cc9f0", "#4361ee", "#9b5de5", "#f15bb5", "#8d5524", "#adb5bd"];
export const BRUSH = [5, 12, 26];

function pathOf(s) {
  let d = `M${s[2]} ${s[3]}`;
  if (s.length === 4) return d + "h0.1";
  for (let i = 4; i < s.length; i += 2) d += `L${s[i]} ${s[i + 1]}`;
  return d;
}

export function drawingSvg(strokes, cls = "") {
  return h("svg", { viewBox: "0 0 1000 1000", class: "paper " + cls, "aria-hidden": "true" },
    (strokes || []).map((s) => h("path", { d: pathOf(s), stroke: INK[s[0]] || INK[0], "stroke-width": s[1], fill: "none", "stroke-linecap": "round", "stroke-linejoin": "round" })));
}

export class DrawPad {
  constructor(canvas, onChange) {
    this.cv = canvas;
    this.ctx = canvas.getContext("2d");
    this.strokes = [];
    this.color = 0;
    this.width = 1;
    this.onChange = onChange || (() => {});
    this.cur = null;
    this.size();
    new ResizeObserver(() => this.size()).observe(canvas);
    canvas.addEventListener("pointerdown", (e) => this.down(e));
    canvas.addEventListener("pointermove", (e) => this.move(e));
    ["pointerup", "pointercancel", "pointerleave"].forEach((n) => canvas.addEventListener(n, () => this.up()));
  }

  size() {
    const r = this.cv.getBoundingClientRect();
    const dpr = window.devicePixelRatio || 1;
    this.cv.width = Math.max(1, Math.round(r.width * dpr));
    this.cv.height = Math.max(1, Math.round(r.height * dpr));
    this.redraw();
  }

  pt(e) {
    const r = this.cv.getBoundingClientRect();
    const c = (v) => Math.min(1000, Math.max(0, Math.round(v)));
    return [c(((e.clientX - r.left) / r.width) * 1000), c(((e.clientY - r.top) / r.height) * 1000)];
  }

  down(e) {
    e.preventDefault();
    this.cv.setPointerCapture(e.pointerId);
    const [x, y] = this.pt(e);
    this.cur = [this.color, BRUSH[this.width], x, y];
    this.strokes.push(this.cur);
    this.redraw();
  }

  move(e) {
    if (!this.cur) return;
    const [x, y] = this.pt(e);
    const n = this.cur.length;
    if (Math.hypot(x - this.cur[n - 2], y - this.cur[n - 1]) < 7) return;
    this.cur.push(x, y);
    this.redraw();
  }

  up() {
    if (!this.cur) return;
    this.cur = null;
    this.onChange(this);
  }

  undo() { this.strokes.pop(); this.redraw(); this.onChange(this); }
  clear() { this.strokes = []; this.redraw(); this.onChange(this); }

  redraw() {
    const { ctx, cv } = this;
    const k = cv.width / 1000;
    ctx.fillStyle = "#fffdf7";
    ctx.fillRect(0, 0, cv.width, cv.height);
    ctx.lineCap = ctx.lineJoin = "round";
    for (const s of this.strokes) {
      ctx.strokeStyle = ctx.fillStyle = INK[s[0]];
      ctx.lineWidth = s[1] * k;
      ctx.beginPath();
      ctx.moveTo(s[2] * k, s[3] * k);
      if (s.length === 4) ctx.lineTo(s[2] * k + 0.1, s[3] * k);
      for (let i = 4; i < s.length; i += 2) ctx.lineTo(s[i] * k, s[i + 1] * k);
      ctx.stroke();
    }
  }
}
