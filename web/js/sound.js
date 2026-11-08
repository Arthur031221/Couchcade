// Short synthesized blips, no audio files. The TV is the only page that plays them.
let ctx = null;
let muted = localStorage.getItem("cc.muted") === "1";

export const isMuted = () => muted;
export function setMuted(m) {
  muted = m;
  localStorage.setItem("cc.muted", m ? "1" : "0");
}

function tone(freq, start, dur, type = "sine", vol = 0.08) {
  if (muted) return;
  ctx = ctx || new (window.AudioContext || window.webkitAudioContext)();
  const o = ctx.createOscillator();
  const g = ctx.createGain();
  o.type = type;
  o.frequency.value = freq;
  g.gain.setValueAtTime(vol, ctx.currentTime + start);
  g.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + start + dur);
  o.connect(g).connect(ctx.destination);
  o.start(ctx.currentTime + start);
  o.stop(ctx.currentTime + start + dur);
}

export const sfx = {
  join: () => { tone(523, 0, 0.12); tone(784, 0.1, 0.16); },
  tick: () => tone(880, 0, 0.05, "square", 0.04),
  ok: () => { tone(659, 0, 0.1); tone(988, 0.09, 0.18); },
  go: () => { tone(440, 0, 0.1, "sawtooth"); tone(880, 0.1, 0.3, "sawtooth"); },
  reveal: () => { tone(392, 0, 0.12); tone(523, 0.12, 0.12); tone(659, 0.24, 0.25); },
  win: () => [523, 659, 784, 1047].forEach((f, i) => tone(f, i * 0.13, 0.25, "triangle", 0.1)),
};
