// WebSocket link with automatic reconnect. In the hosted demo the same protocol
// runs against the engine compiled to WebAssembly, found on the parent page.
function demoHost() {
  try {
    return window.parent !== window && window.parent.__couchEngine ? window.parent.__couchEngine : null;
  } catch {
    return null;
  }
}

export class Link {
  constructor({ onmessage, onstatus, hello }) {
    this.onmessage = onmessage;
    this.onstatus = onstatus;
    this.hello = hello;
    this.tries = 0;
    this.closed = false;
    this.pending = [];
  }

  connect() {
    if (this.closed) return;
    this.onstatus(this.tries ? "reconnecting" : "connecting");
    const demo = demoHost();
    if (demo) {
      this.sock = demo.open((data) => this.onmessage(JSON.parse(data)));
      this.ready();
      return;
    }
    const proto = location.protocol === "https:" ? "wss" : "ws";
    const dir = location.pathname.replace(/[^/]*$/, "");
    const ws = new WebSocket(`${proto}://${location.host}${dir}ws`);
    this.ws = ws;
    ws.onopen = () => {
      this.sock = { send: (s) => ws.send(s), close: () => ws.close() };
      this.ready();
    };
    ws.onmessage = (e) => this.onmessage(JSON.parse(e.data));
    ws.onclose = () => {
      this.sock = null;
      if (this.closed) return;
      this.onstatus("reconnecting");
      this.tries++;
      setTimeout(() => this.connect(), Math.min(5000, 400 * 2 ** Math.min(this.tries, 4)));
    };
  }

  ready() {
    this.tries = 0;
    this.onstatus("open");
    this.send(this.hello());
    this.pending.splice(0).forEach((m) => this.send(m));
  }

  send(obj) {
    const s = JSON.stringify(obj);
    if (this.sock) this.sock.send(s);
  }

  close() {
    this.closed = true;
    if (this.sock) this.sock.close();
  }
}
