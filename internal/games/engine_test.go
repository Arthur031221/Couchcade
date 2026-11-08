package games_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Arthur031221/Couchcade/internal/engine"
)

func TestReconnectKeepsIdentityAndScore(t *testing.T) {
	h := newHarness(t, 2)
	h.start("echo")
	h.act(0, engine.Action{K: "submit", S: "a"})
	h.act(1, engine.Action{K: "submit", S: "a"})
	h.waitPhase("reveal")
	tok := h.conns[0].find("joined")["token"].(string)
	id := h.conns[0].find("joined")["id"]
	h.m.Close(h.phones[0])
	c := &fakeConn{}
	s := engine.NewSession(c)
	h.m.Handle(s, []byte(fmt.Sprintf(`{"t":"join","code":%q,"token":%q}`, h.code, tok)))
	if c.find("joined")["id"] != id {
		t.Fatal("a returning phone must get its old player back")
	}
	h.phones[0], h.conns[0] = s, c
	h.step(0)
	if h.score(0) != 100 {
		t.Fatalf("score lost on reconnect: %d", h.score(0))
	}
}

func TestKickMuteAndNames(t *testing.T) {
	h := newHarness(t, 0)
	a := h.join("Sam")
	b := h.join("Sam")
	room, _ := h.tvState()
	names := room["players"].([]any)
	if names[0].(map[string]any)["name"] != "Sam" || names[1].(map[string]any)["name"] != "Sam 2" {
		t.Fatalf("duplicate names not separated: %v", names)
	}
	id := h.conns[b].find("joined")["id"]
	h.host("kick", fmt.Sprintf(`,"id":%q`, id))
	if !h.conns[b].closed {
		t.Fatal("kicked phone still connected")
	}
	h.step(0)
	room, _ = h.tvState()
	if len(room["players"].([]any)) != 1 {
		t.Fatal("kicked player still listed")
	}
	_ = a
}

func TestMutedPlayerIsIgnored(t *testing.T) {
	h := newHarness(t, 3)
	id := h.conns[0].find("joined")["id"]
	h.host("mute", fmt.Sprintf(`,"id":%q`, id))
	h.start("echo")
	h.act(0, engine.Action{K: "submit", S: "x"})
	h.act(1, engine.Action{K: "submit", S: "x"})
	h.act(2, engine.Action{K: "submit", S: "x"})
	h.waitPhase("reveal")
	if h.score(0) != 0 || h.score(1) != 100 {
		t.Fatalf("muted player counted: %d %d", h.score(0), h.score(1))
	}
}

func TestOnlyHostAndTVCanStart(t *testing.T) {
	h := newHarness(t, 3)
	h.m.Handle(h.phones[1], []byte(`{"t":"host","op":"start","game":"echo"}`))
	h.step(0)
	if h.roomPhase() != "lobby" {
		t.Fatal("a non-host phone started a game")
	}
	h.m.Handle(h.phones[0], []byte(`{"t":"host","op":"start","game":"echo"}`))
	h.step(0)
	if h.roomPhase() != "game" {
		t.Fatal("the host phone could not start a game")
	}
}

func TestNeedEnoughPlayers(t *testing.T) {
	h := newHarness(t, 2)
	h.start("mole")
	if h.tvc.find("err") == nil || h.roomPhase() != "lobby" {
		t.Fatal("mole needs three players")
	}
}

func TestSwitchLanguageChangesPrompts(t *testing.T) {
	h := newHarness(t, 2)
	h.host("lang", `,"lang":"ja"`)
	h.start("echo")
	p := h.tvg()["prompt"].(string)
	if len(p) == len([]rune(p)) { // ASCII only means we still got English
		t.Fatalf("expected Japanese prompt, got %q", p)
	}
}

func TestIdleRoomsAreSwept(t *testing.T) {
	h := newHarness(t, 1)
	h.m.Close(h.tv)
	h.step(0)
	h.now = h.now.Add(31 * time.Minute)
	h.m.TickAll()
	h.m.Sweep()
	if h.m.RoomCount() != 1 {
		t.Fatal("a room with a connected phone must be kept")
	}
	h.m.Close(h.phones[0])
	h.m.Sweep()
	if h.m.RoomCount() != 0 {
		t.Fatal("abandoned room kept")
	}
}

func TestSecondJoinOnOneConnectionIsIgnored(t *testing.T) {
	h := newHarness(t, 1)
	h.m.Handle(h.phones[0], []byte(fmt.Sprintf(`{"t":"join","code":%q,"name":"Ghost"}`, h.code)))
	room, _ := h.tvState()
	if n := len(room["players"].([]any)); n != 1 {
		t.Fatalf("want 1 player, got %d", n)
	}
}

func TestOneTVConnectionOwnsOneRoom(t *testing.T) {
	h := newHarness(t, 0)
	for i := 0; i < 50; i++ {
		h.m.Handle(h.tv, []byte(`{"t":"tv","code":"ZZZZ","token":"wrong"}`))
	}
	if n := h.m.RoomCount(); n != 1 {
		t.Fatalf("a TV that keeps reconnecting left %d rooms", n)
	}
}

func TestNamesWithInvisibleCharactersAreCleaned(t *testing.T) {
	h := newHarness(t, 0)
	h.join("Bob")
	h.join("Bob\u200b")
	room, _ := h.tvState()
	names := map[string]bool{}
	for _, p := range room["players"].([]any) {
		names[p.(map[string]any)["name"].(string)] = true
	}
	if len(names) != 2 || names["Bob\u200b"] {
		t.Fatalf("names %v", names)
	}
}

func TestPhaseWaitsForTheClockWhenEveryoneIsGone(t *testing.T) {
	h := newHarness(t, 2)
	h.start("echo")
	h.m.Close(h.phones[0])
	h.m.Close(h.phones[1])
	h.step(2 * time.Second)
	if h.phase() != "answer" {
		t.Fatalf("game skipped ahead with nobody connected, now in %q", h.phase())
	}
}
