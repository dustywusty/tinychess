// Tiny Chess: Go server with SSE + vanilla JS UI
// Features:
// - Shareable URL /{uuid} for each game
// - Server is the source of truth; validates moves with notnil/chess
// - Live updates via Server-Sent Events (EventSource)
// - Simple UI (Unicode pieces, click-to-move)
// - Reset; copy-link; promotions (auto-queen)
// - Theme picker (accent + light/dark), emoji reactions
// - Captured pieces (by White / by Black) with localStorage persistence
// - Coordinates (a–h / 1–8) on the board
// - PGN (SAN) lines + UCI move list (from→to)
// - NEW: Home remembers recent games (localStorage), prevents new-game if one is active,
//        and shows results (1-0/0-1/½-½) when finished

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/notnil/chess"
)

// ----- Game hub -----

type hub struct {
	mu    sync.Mutex
	games map[string]*game
}

type game struct {
	mu        sync.Mutex
	g         *chess.Game
	watchers  map[chan []byte]struct{}
	lastReact map[string]time.Time
	lastSeen  time.Time
}

func newHub() *hub { return &hub{games: make(map[string]*game)} }

func (h *hub) get(id string) *game {
	h.mu.Lock()
	defer h.mu.Unlock()
	if g, ok := h.games[id]; ok {
		return g
	}
	ng := &game{g: chess.NewGame(), watchers: make(map[chan []byte]struct{})}
	h.games[id] = ng
	return ng
}

func (g *game) touch() {
	g.mu.Lock()
	g.lastSeen = time.Now()
	g.mu.Unlock()
}

// Build a UCI (from->to) list by replaying moves on a temp game,
// ensuring we pass the correct position to UCINotation{}.Encode.
func (g *game) movesUCI() []string {
	ms := g.g.Moves()
	out := make([]string, 0, len(ms))
	tmp := chess.NewGame()
	uci := chess.UCINotation{}
	for _, m := range ms {
		// Encode using the position before the move
		s := uci.Encode(tmp.Position(), m)
		out = append(out, s)
		// Advance tmp using the encoded UCI for tmp's position
		if mv2, err := uci.Decode(tmp.Position(), s); err == nil {
			_ = tmp.Move(mv2)
		}
	}
	return out
}

func (g *game) stateLocked() map[string]any {
	pos := g.g.Position()
	fen := pos.String()
	turn := pos.Turn().String()
	status := ""
	if g.g.Outcome() != chess.NoOutcome {
		status = fmt.Sprintf("%s by %s", g.g.Outcome().String(), g.g.Method().String())
	}
	pgn := g.g.String()
	return map[string]any{
		"kind":     "state",
		"fen":      fen,
		"turn":     turn,
		"status":   status,
		"pgn":      pgn,
		"uci":      g.movesUCI(),
		"lastSeen": g.lastSeen.UnixMilli(),
	}
}

func (g *game) broadcast() {
	g.mu.Lock()
	state := g.stateLocked()
	data, _ := json.Marshal(state)
	for ch := range g.watchers {
		select { // don't block a slow client
		case ch <- data:
		default:
		}
	}
	g.mu.Unlock()
}

// ----- HTTP -----

var H = newHub()

func main() {
	http.HandleFunc("/new", handleNew)
	http.HandleFunc("/sse/", handleSSE)
	http.HandleFunc("/move/", handleMove)
	http.HandleFunc("/react/", handleReact)
	http.HandleFunc("/reset/", handleReset)
	http.HandleFunc("/", handlePage)

	log.Printf("Tiny Chess listening on http://localhost:8080 …")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleNew(w http.ResponseWriter, r *http.Request) {
	id := uuid.NewString()
	http.Redirect(w, r, "/"+id, http.StatusFound)
}

func handlePage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" || path == "index.html" {
		ioWriteHTML(w, homeHTML)
		return
	}
	_ = H.get(path)
	ioWriteHTML(w, strings.ReplaceAll(gameHTML, "{{GAME_ID}}", path))
}

func handleSSE(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/sse/")
	g := H.get(id)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan []byte, 16)
	g.mu.Lock()
	g.watchers[ch] = struct{}{}
	initial, _ := json.Marshal(g.stateLocked())
	g.mu.Unlock()

	fmt.Fprintf(w, "data: %s\n\n", initial)
	flusher.Flush()

	g.touch()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			g.mu.Lock()
			delete(g.watchers, ch)
			g.mu.Unlock()
			return
		case msg := <-ch:
			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(msg)
			_, _ = w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}

func handleMove(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/move/")
	g := H.get(id)

	var m struct {
		UCI string `json:"uci"`
	}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json"})
		return
	}
	uci := strings.ToLower(strings.TrimSpace(m.UCI))
	// default to queen on bare promotion UCI
	if len(uci) == 4 && isPromotionToLastRank(uci) {
		uci += "q"
	}

	g.mu.Lock()
	move, err := chess.UCINotation{}.Decode(g.g.Position(), uci)
	if err == nil {
		err = g.g.Move(move)
	}
	state := g.stateLocked()
	g.mu.Unlock()

	g.touch()

	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error(), "state": state})
		return
	}
	go g.broadcast()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "state": state})
}

// Emoji reactions with 5s/IP cooldown
func handleReact(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/react/")
	g := H.get(id)

	var body struct {
		Emoji  string `json:"emoji"`
		Sender string `json:"sender"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json"})
		return
	}
	allowed := map[string]struct{}{
		"👍": {}, "👎": {}, "❤️": {}, "😠": {}, "😢": {}, "🎉": {}, "👏": {},
		"😂": {}, "🤣": {}, "😎": {}, "🤔": {}, "😏": {}, "🙃": {}, "😴": {}, "🫡": {}, "🤯": {}, "🤡": {},
		"♟️": {}, "♞": {}, "♝": {}, "♜": {}, "♛": {}, "♚": {}, "⏱️": {}, "🏳️": {}, "🔄": {}, "🏆": {},
		"🔥": {}, "💀": {}, "🩸": {}, "⚡": {}, "🚀": {}, "🕳️": {}, "🎯": {}, "💥": {}, "🧠": {},
		"🍿": {}, "☕": {}, "🐢": {}, "🐇": {}, "🤝": {}, "🤬": {},
		"🪦": {}, "🐌": {}, "🎭": {}, "🐙": {}, "🦄": {}, "🐒": {},
	}
	if _, ok := allowed[body.Emoji]; !ok {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "unsupported emoji"})
		return
	}

	ip := clientIP(r)
	now := time.Now()

	g.mu.Lock()
	if g.lastReact == nil {
		g.lastReact = make(map[string]time.Time)
	}
	if t, ok := g.lastReact[ip]; ok && now.Sub(t) < 5*time.Second {
		wait := int(5 - now.Sub(t).Seconds())
		g.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": fmt.Sprintf("cooldown %ds", wait)})
		return
	}
	g.lastReact[ip] = now

	payload := map[string]any{
		"kind":   "emoji",
		"emoji":  body.Emoji,
		"at":     now.UnixMilli(),
		"sender": body.Sender,
	}
	data, _ := json.Marshal(payload)
	for ch := range g.watchers {
		select {
		case ch <- data:
		default:
		}
	}
	g.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/reset/")
	g := H.get(id)
	g.mu.Lock()
	g.g = chess.NewGame()
	state := g.stateLocked()
	g.mu.Unlock()
	go g.broadcast()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "state": state})
}

// ----- Helpers -----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func isPromotionToLastRank(uci string) bool {
	if len(uci) != 4 {
		return false
	}
	to := uci[2:]
	return to[1] == '1' || to[1] == '8'
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func ioWriteHTML(w http.ResponseWriter, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// prevent stale HTML during rapid iteration
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

// ----- HTML (home + game) -----

const homeHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>Tiny Chess</title>
<style>
  :root { --accent:#6ee7ff; --ok:#22c55e; --err:#ef4444; }
  :root,[data-theme="dark"] {
    /* Accent-tinted theme (dark) */
    --bg:     color-mix(in oklab, var(--accent) 6%,  #0b0d11);
    --panel:  color-mix(in oklab, var(--accent) 10%, #141821);
    --text:#e5e7eb;
    /* Buttons */
    --btn-bg:#1a2230; --btn-hover:#1f2a3a; --btn-text:#e5e7eb; --btn-border:#2a3345;
  }
  [data-theme="light"] {
    /* Accent-tinted theme (light) */
    --bg:     color-mix(in oklab, var(--accent) 8%,  #f7f7fb);
    --panel:  color-mix(in oklab, var(--accent) 12%, #ffffff);
    --text:#0f172a;
    /* Buttons */
    --btn-bg:    color-mix(in oklab, var(--accent) 14%, white);
    --btn-hover: color-mix(in oklab, var(--accent) 22%, white);
    --btn-text:  #0f172a;
    --btn-border: color-mix(in oklab, var(--accent) 30%, #b6c3d9);
  }

  * { box-sizing: border-box; }
  body { margin:0; background:var(--bg); color:var(--text);
         font:14px/1.4 system-ui,-apple-system,Segoe UI,Roboto,Ubuntu,Cantarell,Noto Sans,sans-serif; }

  header { padding:10px 14px; display:flex; gap:8px; align-items:center;
           border-bottom:1px solid var(--btn-border); background:var(--panel); position:sticky; top:0; }
  .title { font-weight:600; letter-spacing:0.2px; }
  .btn{
    cursor:pointer; border:1px solid var(--btn-border);
    background:var(--btn-bg); color:var(--btn-text);
    border-radius:10px; padding:8px 12px; font-weight:600;
  }
  .btn:hover{ background:var(--btn-hover); }
  .btn:focus-visible{ outline:2px solid var(--accent); outline-offset:2px; border-color:transparent; }

  .theme { display:flex; gap:6px; align-items:center; }
  .swatch,.mode { border:1px solid var(--btn-border); }
  .swatch { width:16px; height:16px; border-radius:999px; cursor:pointer; }
  .mode   { width:16px; height:16px; border-radius:4px;   cursor:pointer; }
  .active { outline:2px solid var(--accent); outline-offset:2px; }

  main { max-width:800px; margin:40px auto; padding:0 16px; text-align:center; }
  h1 { font-weight:700; margin-bottom:12px; }
  p  { opacity:.85; }
  footer { opacity:0.7; padding:8px 14px 24px; text-align:center; }

  /* Recent list */
  .recent { max-width:800px; margin:24px auto; padding:0 16px; text-align:left; }
  .card { background:var(--panel); border:1px solid var(--btn-border); border-radius:12px; padding:12px; margin:10px 0; }
  .row { display:flex; gap:8px; align-items:center; flex-wrap:wrap; }
  .mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; }
  .pill { display:inline-block; border:1px solid var(--btn-border); padding:2px 6px; border-radius:999px; font-size:12px; opacity:.9; }
</style>
</head>
<body>
  <header>
    <div class="title">♟️ Tiny Chess</div>
    <div style="flex:1"></div>
    <div class="theme" id="themectl">
      <button class="swatch" data-accent="#6ee7ff" style="background:#6ee7ff" aria-label="Accent cyan"></button>
      <button class="swatch" data-accent="#a78bfa" style="background:#a78bfa" aria-label="Accent purple"></button>
      <button class="swatch" data-accent="#f472b6" style="background:#f472b6" aria-label="Accent pink"></button>
      <button class="swatch" data-accent="#f59e0b" style="background:#f59e0b" aria-label="Accent amber"></button>
      <button class="swatch" data-accent="#10b981" style="background:#10b981" aria-label="Accent green"></button>
      <button class="mode" data-theme="light" style="background:#ffffff" aria-label="Light mode"></button>
      <button class="mode" data-theme="dark"  style="background:#000000" aria-label="Dark mode"></button>
    </div>
    <a class="btn" href="/new" id="newgame">New game</a>
  </header>

  <main>
    <h1>Play chess with a link</h1>
    <p>Click “New game” to create a shareable URL. Anyone with the link can watch and move.</p>
    <p><a class="btn" href="/new" id="newgame2">New game</a></p>
  </main>

  <section class="recent">
    <h2>Recent games (this browser)</h2>
    <div id="recent"></div>
  </section>

  <footer>Built with Go, SSE & vanilla JS — with ❤️ by Dusty and his well-treated robots.</footer>
  <script defer data-domain="tinychess.bitchimfabulo.us" src="https://plausible.io/js/script.outbound-links.js"></script>
<script>
(function(){
  const root = document.documentElement;
  let theme  = localStorage.getItem('theme')  || 'dark';
  let accent = localStorage.getItem('accent') ||
               getComputedStyle(root).getPropertyValue('--accent').trim() || '#6ee7ff';
  root.setAttribute('data-theme', theme);
  root.style.setProperty('--accent', accent);

  function markActive(){
    document.querySelectorAll('.swatch').forEach(el=>{
      el.classList.toggle('active', el.getAttribute('data-accent')===accent);
    });
    document.querySelectorAll('.mode').forEach(el=>{
      el.classList.toggle('active', el.getAttribute('data-theme')===theme);
    });
  }
  markActive();

  document.addEventListener('click', (e)=>{
    const t = e.target;
    if (t.matches('.swatch')){
      accent = t.getAttribute('data-accent');
      root.style.setProperty('--accent', accent);
      localStorage.setItem('accent', accent);
      markActive();
    } else if (t.matches('.mode')){
      theme = t.getAttribute('data-theme');
      root.setAttribute('data-theme', theme);
      localStorage.setItem('theme', theme);
      markActive();
    }
  });

  // ----- Recent/active games -----
  const KEY = 'tinychess:games:v1';
  function loadGames(){ try { return JSON.parse(localStorage.getItem(KEY) || '{}'); } catch { return {}; } }
  function saveGames(map){ try { localStorage.setItem(KEY, JSON.stringify(map)); } catch {} }
  function byLastSeenDesc(a,b){ return (b.lastSeen||0)-(a.lastSeen||0); }
  function hasResult(g){ return !!(g && g.result); }
  function activeGames(){ const m = loadGames(); return Object.values(m).filter(function(g){ return !hasResult(g); }); }

  function renderRecent(){
    const box = document.getElementById('recent'); 
    if(!box) return;

    const games = Object.values(loadGames()).sort(byLastSeenDesc);
    if (!games.length){
      box.innerHTML = '<p style="opacity:.8">No games yet — start one above.</p>';
      return;
    }

    box.innerHTML = '';
    for (var i=0;i<games.length;i++){
      var g = games[i];
      var a = document.createElement('div');
      a.className = 'card';

      // prefer server lastSeen, fallback to local, then createdAt
      var seen = g.lastSeen || g.lastSeenLocal || g.createdAt || Date.now();
      var when = new Date(seen).toLocaleString();

      var res  = g.result ? '<span class="pill">Result: ' + g.result + '</span>' : '<span class="pill">In progress</span>';
      var stat = g.status ? '<span class="pill">' + g.status + '</span>' : '';

      a.innerHTML =
        '<div class="row">' +
        '  <strong>ID:</strong> <span class="mono">' + g.id + '</span> ' +
        '  ' + res + ' ' + stat +
        '</div>' +
        '<div class="row" style="margin-top:6px;">' +
        '  <button class="btn" data-goto="' + g.id + '">Open</button>' +
        '  <button class="btn" data-copy="' + g.id + '">Copy link</button>' +
        '  <button class="btn" data-remove="' + g.id + '">Forget</button>' +
        '  <span style="opacity:.7; margin-left:auto;">Last seen: ' + when + '</span>' +
        '</div>';
      box.appendChild(a);
    }
  }

  renderRecent();

  document.addEventListener('click', function(e){
    const t = e.target;
    if (t.matches('[data-goto]')){ location.href = '/' + t.getAttribute('data-goto'); }
    if (t.matches('[data-copy]')){ try{ navigator.clipboard.writeText(location.origin + '/' + t.getAttribute('data-copy')); }catch(e){} }
    if (t.matches('[data-remove]')){
      const id = t.getAttribute('data-remove');
      const m = loadGames(); delete m[id]; saveGames(m); renderRecent();
    }
  });

  function handleNewClick(ev){
    const act = activeGames();
    if (act.length){
      act.sort(byLastSeenDesc);
      location.href = '/' + act[0].id;
      ev.preventDefault();
    }
  }
  ['newgame','newgame2'].forEach(function(id){
    const el = document.getElementById(id);
    if (el) el.addEventListener('click', handleNewClick);
  });
})();
</script>
</body>
</html>`

const gameHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>Tiny Chess</title>
<style>
  :root { --accent:#6ee7ff; --ok:#22c55e; --err:#ef4444; }
  :root,[data-theme="dark"] {
    --bg:     color-mix(in oklab, var(--accent) 6%,  #0b0d11);
    --panel:  color-mix(in oklab, var(--accent) 10%, #141821);
    --text:#e5e7eb;
    --piece-light:#111827; --piece-dark:#eef2f7;
    /* Board colors derived from accent for dark theme */
    --sq1: color-mix(in oklab, var(--accent) 18%, white);
    --sq2: color-mix(in oklab, var(--accent) 62%, black);
    --sq3: color-mix(in oklab, var(--accent) 24%, black);
    /* Buttons (dark) */
    --btn-bg:#1a2230; --btn-hover:#1f2a3a; --btn-text:#e5e7eb; --btn-border:#2a3345;
  }
  [data-theme="light"] {
    --bg:     color-mix(in oklab, var(--accent) 8%,  #f7f7fb);
    --panel:  color-mix(in oklab, var(--accent) 12%, #ffffff);
    --text:#0f172a;
    --piece-light:#0f172a; --piece-dark:#0f172a;
    /* Softer board colors for light theme */
    --sq1: color-mix(in oklab, var(--accent) 8%,  white);
    --sq2: color-mix(in oklab, var(--accent) 28%, #7f99b7);
    --sq3: color-mix(in oklab, var(--accent) 14%, #b9cce1);
    /* Buttons (light) */
    --btn-bg:    color-mix(in oklab, var(--accent) 14%, white);
    --btn-hover: color-mix(in oklab, var(--accent) 22%, white);
    --btn-text:  #0f172a;
    --btn-border: color-mix(in oklab, var(--accent) 30%, #b6c3d9);
  }

  * { box-sizing: border-box; }
  body { margin:0; background:var(--bg); color:var(--text);
         font:14px/1.4 system-ui, -apple-system, Segoe UI, Roboto, Ubuntu, Cantarell, Noto Sans, sans-serif; }
  header { padding:10px 14px; display:flex; gap:8px; align-items:center;
           border-bottom:1px solid var(--btn-border); background:var(--panel); position:sticky; top:0; }
  .title { font-weight:600; letter-spacing:0.2px; }
  .wrap { max-width:1000px; margin:0 auto; padding:16px; display:grid; grid-template-columns: 1fr 320px; gap:16px; }
  .board { width:100%; max-width:640px; aspect-ratio:1/1; border:1px solid #2a3345; border-radius:12px; overflow:hidden;
           user-select:none; background:var(--sq3); display:grid; grid-template-rows: repeat(8, 1fr); }
  .rank  { display:grid; grid-template-columns: repeat(8, 1fr); }
  .cell  { display:flex; align-items:center; justify-content:center; font-size: clamp(22px, 6vw, 54px); position:relative; }
  .light { background: var(--sq1); color: var(--piece-light); }
  .dark  { background: var(--sq2); color: var(--piece-dark); }
  .cell.sel { outline:3px solid var(--accent); outline-offset:-3px; }

  /* Coordinates */
  .coord { position:absolute; pointer-events:none; font-size:12px; line-height:1; opacity:.75; }
  .coord-file { bottom:4px; right:6px; }
  .coord-rank { top:4px; left:6px; }
  .light .coord { color: rgba(15,23,42,.65); }
  .dark  .coord { color: rgba(255,255,255,.78); }

  .panel { background:var(--panel); border:1px solid #2a3345; border-radius:12px; padding:12px; }

  /* Theme-aware buttons */
  .btn{
    cursor:pointer; border:1px solid var(--btn-border);
    background:var(--btn-bg); color:var(--btn-text);
    border-radius:10px; padding:8px 12px; font-weight:600;
  }
  .btn:hover{ background:var(--btn-hover); }
  .btn:focus-visible{ outline:2px solid var(--accent); outline-offset:2px; border-color:transparent; }

  .row { display:flex; gap:8px; align-items:center; flex-wrap:wrap; }
  .status { margin-top:10px; min-height:22px; }
  .mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace; }
  a { color: var(--accent); text-decoration: none; }
  footer { opacity:0.7; padding:8px 14px 24px; text-align:center; }

  .theme { display:flex; gap:6px; align-items:center; margin-right:8px; }
  .swatch,.mode { border:1px solid var(--btn-border); }
  .swatch { width:16px; height:16px; border-radius:999px; cursor:pointer; }
  .mode   { width:16px; height:16px; border-radius:4px;   cursor:pointer; }
  .active { outline:2px solid var(--accent); outline-offset:2px; }

  .caps { min-height:22px; display:flex; gap:4px; flex-wrap:wrap; font-size:20px; line-height:1; }
  .reactions{ display:flex; gap:6px; flex-wrap:wrap; }
  .react{ font-size:18px; background:transparent; border:1px solid var(--btn-border);
          border-radius:8px; padding:4px 6px; cursor:pointer; }
  .react[disabled]{ opacity:.55; cursor:not-allowed; }
  .rx{ margin-top:6px; display:flex; gap:6px; flex-wrap:wrap; min-height:22px; }
  @keyframes pop { from{ transform:scale(.4); opacity:0; } to{ transform:scale(1); opacity:1; } }
  .burst{ animation: pop .28s ease-out; }
  @media (max-width: 860px){ .wrap { grid-template-columns: 1fr; } }
  .big-emoji {
    position: fixed;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%) scale(4);
    font-size: 120px;
    pointer-events: none;
    animation: shrinkFade 1.2s ease-out forwards;
    z-index: 9999;
  }
  @keyframes shrinkFade {
    0%   { transform: translate(-50%, -50%) scale(4); opacity: 1; }
    50%  { transform: translate(-50%, -50%) scale(1.2); opacity: 1; }
    100% { transform: translate(-50%, -50%) scale(0.5); opacity: 0; }
} 
</style>
</head>
<body>
  <header>
    <div class="title"><a href="..">♟️ Tiny Chess</a></div>
    <div style="flex:1"></div>
    <div class="theme" id="themectl">
      <button class="swatch" data-accent="#6ee7ff" style="background:#6ee7ff"></button>
      <button class="swatch" data-accent="#a78bfa" style="background:#a78bfa"></button>
      <button class="swatch" data-accent="#f472b6" style="background:#f472b6"></button>
      <button class="swatch" data-accent="#f59e0b" style="background:#f59e0b"></button>
      <button class="swatch" data-accent="#10b981" style="background:#10b981"></button>
      <button class="mode" data-theme="light" style="background:#ffffff"></button>
      <button class="mode" data-theme="dark"  style="background:#000000"></button>
    </div>
    <button class="btn" id="copy">Copy link</button>
    <a class="btn" href="/new">New game</a>
  </header>

  <div class="wrap">
    <div class="board" id="board" aria-label="Chess board"></div>
    <div class="panel">
      <div class="row"><strong>Game:</strong> <span id="gameid" class="mono"></span></div>
      <div class="row"><strong>Turn:</strong> <span id="turn"></span></div>
      <div class="status" id="status"></div>

      <!-- Captured pieces -->
      <div class="row"><strong>White captured:</strong> <span id="cap_by_white" class="caps"></span></div>
      <div class="row"><strong>Black captured:</strong> <span id="cap_by_black" class="caps"></span></div>

      <!-- Reactions -->
      <div class="rx" id="rx"></div>
      <div class="row"><div class="reactions" id="reactbar" aria-label="Reactions"></div></div>

      <div class="row" style="margin-top:8px"><button class="btn" id="reset">Reset</button></div>

      <details style="margin-top:12px;"><summary>PGN</summary>
        <pre id="pgn" style="white-space:pre-wrap;"></pre>
      </details>

      <details style="margin-top:8px"><summary>Moves (from→to)</summary>
        <pre id="lan" style="white-space:pre-wrap;"></pre>
      </details>

      <p style="margin-top:10px; opacity:.8">Tip: Click one square, then another to move. Promotions auto-queen. Anyone with the link can move.</p>
    </div>
  </div>

  <footer>Built with Go, SSE & vanilla JS — with ❤️ by Dusty and his bots.</footer>
<script defer data-domain="tinychess.bitchimfabulo.us" src="https://plausible.io/js/script.outbound-links.js"></script>
<script>
(function(){
  const CLIENT_ID_KEY = "tinychess:clientId";
  let clientId = localStorage.getItem(CLIENT_ID_KEY);
  if (!clientId) {
    clientId = Math.random().toString(36).slice(2);
    localStorage.setItem(CLIENT_ID_KEY, clientId);
  }
  const idFromServer = "{{GAME_ID}}";
  const gameId = idFromServer || location.pathname.replace(/^\/+/, '');
  const boardEl = document.getElementById('board');
  const statusEl = document.getElementById('status');
  const turnEl = document.getElementById('turn');
  const pgnEl  = document.getElementById('pgn');
  const lanEl  = document.getElementById('lan');
  const gameIdEl = document.getElementById('gameid');
  const capWhiteEl = document.getElementById('cap_by_white');
  const capBlackEl = document.getElementById('cap_by_black');
  gameIdEl.textContent = gameId || '(none)';

  const glyph = {
    'P':'\u2659','N':'\u2658','B':'\u2657','R':'\u2656','Q':'\u2655','K':'\u2654',
    'p':'\u265F','n':'\u265E','b':'\u265D','r':'\u265C','q':'\u265B','k':'\u265A'
  };
  let selected = null;

  // ---- local game index (per-browser) ----
  const INDEX_KEY = 'tinychess:games:v1';
  function loadIndex(){ try { return JSON.parse(localStorage.getItem(INDEX_KEY) || '{}'); } catch { return {}; } }
  function saveIndex(m){ try { localStorage.setItem(INDEX_KEY, JSON.stringify(m)); } catch {} }
  function rememberGame(id){ if (!id) return; var m = loadIndex(); if (!m[id]) m[id] = { id:id, createdAt: Date.now(), lastSeen: Date.now(), moves: 0, result: null, status: '' }; else m[id].lastSeen = Date.now(); saveIndex(m); }
  function setGameState(id, fields){
    if (!id) return;
    var m = loadIndex();
    // Merge fields (server lastSeen will flow through)
    m[id] = Object.assign(m[id] || { id:id, createdAt: Date.now() }, fields);
    saveIndex(m);
  }
  rememberGame(gameId);

  // Theme picker
  const root = document.documentElement;
  let theme = localStorage.getItem('theme') || 'dark';
  let accent = localStorage.getItem('accent') || getComputedStyle(root).getPropertyValue('--accent').trim() || '#6ee7ff';
  root.setAttribute('data-theme', theme);
  root.style.setProperty('--accent', accent);
  function markActive(){
    var sw = document.querySelectorAll('.swatch');
    for (var i = 0; i < sw.length; i++) {
      if (sw[i].getAttribute('data-accent') === accent) sw[i].classList.add('active');
      else sw[i].classList.remove('active');
    }
    var md = document.querySelectorAll('.mode');
    for (var j = 0; j < md.length; j++) {
      if (md[j].getAttribute('data-theme') === theme) md[j].classList.add('active');
      else md[j].classList.remove('active');
    }
  }
  markActive();
  document.addEventListener('click', (e)=>{
    const t = e.target;
    if (t.matches('.swatch')){ accent = t.getAttribute('data-accent'); root.style.setProperty('--accent', accent); localStorage.setItem('accent', accent); markActive(); }
    else if (t.matches('.mode')){ theme = t.getAttribute('data-theme'); root.setAttribute('data-theme', theme); localStorage.setItem('theme', theme); markActive(); }
  });

  // Reactions
  const rxEl = document.getElementById('rx');
  const reactBar = document.getElementById('reactbar');
  const EMOJIS = [
    "👍","👎","❤️","😠","😢","🎉","👏",
    "😂","🤣","😎","🤔","😏","🙃","😴","🫡","🤯","🤡",
    "♟️","♞","♝","♜","♛","♚","⏱️","🏳️","🔄","🏆",
    "🔥","💀","🩸","⚡","🚀","🕳️","🎯","💥","🧠",
    "🍿","☕","🐢","🐇","🤝","🤬",
    "🪦","🐌","🎭","🐙","🦄","🐒"
  ];
  const COOLDOWN_MS = 5000; let lastReact = 0;
  function buildReactBar(){ if(!reactBar) return; reactBar.innerHTML=''; EMOJIS.forEach(function(e){ var b=document.createElement('button'); b.className='react'; b.textContent=e; b.title='Send reaction'; b.addEventListener('click', function(){ sendReaction(e,b); }); reactBar.appendChild(b); }); }
  function showReaction(e){
    // Big flash
    const big = document.createElement('div');
    big.textContent = e;
    big.className = 'big-emoji';
    document.body.appendChild(big);
    setTimeout(() => big.remove(), 1000);

    // Small burst under reactions bar
    if (rxEl){
      const small = document.createElement('span');
      small.textContent = e;
      small.className = 'burst';
      rxEl.appendChild(small);
      setTimeout(() => small.remove(), 1600);
    }
  }

  async function sendReaction(emoji, btn){
    if(!gameId) return;
    const now = Date.now();
    if (now - lastReact < COOLDOWN_MS) {
      if (btn) { btn.style.background = 'var(--err)'; setTimeout(()=>btn.style.background='',600); }
      status('Hold up… cooldown', true);
      return;
    }
    lastReact = now;
    if (btn){ btn.disabled = true; setTimeout(()=>btn.disabled = false, COOLDOWN_MS); }

    try{
      const res = await fetch('/react/' + gameId, {
        method:'POST',
        headers:{'Content-Type':'application/json'},
        body: JSON.stringify({emoji:emoji, sender: clientId})
      });
      const j = await res.json();
      if (!j.ok){
        if (btn) { btn.style.background = 'var(--err)'; setTimeout(()=>btn.style.background='',600); }
        status(j.error || 'reaction failed', true);
      } else {
        if (btn) { btn.style.background = 'var(--ok)'; setTimeout(()=>btn.style.background='',600); }
        // no local burst here — SSE will handle for others
      }
    }catch(_){
      if (btn) { btn.style.background = 'var(--err)'; setTimeout(()=>btn.style.background='',600); }
    }
  }

  buildReactBar();

  // --- board helpers ---
  function cellSquare(row, col) {
    const file = String.fromCharCode('a'.charCodeAt(0) + col);
    const rank = String(8 - row);
    return file + rank;
  }

  function renderFEN(fen){
    const board = fen.split(' ')[0].split('/');
    boardEl.innerHTML = '';

    for (let r = 0; r < 8; r++) {
      const row = document.createElement('div'); row.className = 'rank';
      const fenRank = board[r];
      const cells = [];

      for (let i = 0; i < fenRank.length; i++) {
        const ch = fenRank[i];
        if (/\d/.test(ch)) {
          const n = parseInt(ch, 10);
          for (let k = 0; k < n; k++) cells.push('');
        } else {
          cells.push(ch);
        }
      }

      for (let c = 0; c < 8; c++) {
        const piece = cells[c] || '';
        const cell  = document.createElement('div');
        cell.className = 'cell ' + (((r + c) % 2 === 1) ? 'light' : 'dark'); // a8 dark
        const sq = cellSquare(r, c);
        cell.dataset.square = sq;
        cell.textContent = glyph[piece] || '';

        // coordinates
        if (r === 7) {
          const f = document.createElement('span');
          f.className = 'coord coord-file';
          f.textContent = String.fromCharCode(97 + c);
          cell.appendChild(f);
        }
        if (c === 0) {
          const rr = document.createElement('span');
          rr.className = 'coord coord-rank';
          rr.textContent = String(8 - r);
          cell.appendChild(rr);
        }

        if (selected && sq === selected) cell.classList.add('sel');
        row.appendChild(cell);
      }
      boardEl.appendChild(row);
    }
  }

  function renderSelected(){
    document.querySelectorAll('.cell')
      .forEach(function(el){ el.classList.toggle('sel', el.dataset.square === selected); });
  }

  async function makeMove(uci){
    if(!gameId){ status('No game id'); return; }
    try{
      const res = await fetch('/move/' + gameId, { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({uci:uci}) });
      const j = await res.json();
      if(!j.ok){ status('Illegal move: '+(j.error||'unknown'), true); }
    }catch(err){ status('Network error', true); }
  }

  // Board-level click handler
  boardEl.addEventListener('click', (e) => {
    const rect = boardEl.getBoundingClientRect();
    const x = Math.min(Math.max(0, e.clientX - rect.left), rect.width  - 0.01);
    const y = Math.min(Math.max(0, e.clientY - rect.top),  rect.height - 0.01);
    const col = Math.floor((x / rect.width)  * 8);
    const row = Math.floor((y / rect.height) * 8);
    const sq  = cellSquare(row, col);

    if (!selected) { selected = sq; renderSelected(); return; }
    if (selected === sq) { selected = null; renderSelected(); return; }
    const uci = (selected + sq).toLowerCase();
    selected = null; renderSelected();
    makeMove(uci);
  });

  function status(msg, isErr){ statusEl.textContent = msg || ''; statusEl.style.color = isErr? 'var(--err)' : 'inherit'; }

  // ---- Captured pieces (derived from FEN) + persisted per game ----
  var startCounts = {P:8,N:2,B:2,R:2,Q:1,K:1,p:8,n:2,b:2,r:2,q:1,k:1};

  function countsFromFEN(fen){
    var boardOnly = fen.split(' ')[0];
    var c = {P:0,N:0,B:0,R:0,Q:0,K:0,p:0,n:0,b:0,r:0,q:0,k:0};
    for (var i=0;i<boardOnly.length;i++){
      var ch = boardOnly[i];
      if (/[prnbqkPRNBQK]/.test(ch)) c[ch] = (c[ch]||0)+1;
    }
    return c;
  }

  function capturedFromFEN(fen){
    var cur = countsFromFEN(fen);

    var lostWhite = {P:0,N:0,B:0,R:0,Q:0,K:0};
    for (var k in lostWhite){
      lostWhite[k] = Math.max(0, (startCounts[k]||0) - (cur[k]||0));
    }

    var lostBlack = {p:0,n:0,b:0,r:0,q:0,k:0};
    for (var k2 in lostBlack){
      lostBlack[k2] = Math.max(0, (startCounts[k2]||0) - (cur[k2]||0));
    }

    var byWhite = [];
    var byBlack = [];

    for (var k3 in lostBlack){
      for (var i=0;i<lostBlack[k3];i++) byWhite.push(glyph[k3]);
    }
    for (var k4 in lostWhite){
      for (var j=0;j<lostWhite[k4];j++) byBlack.push(glyph[k4]);
    }
    return {byWhite: byWhite, byBlack: byBlack};
  }

  // --- formatting helpers ---
  function formatPGNLines(pgn) {
    if (!pgn) return '';
    const tokens = pgn.trim().split(/\s+/);
    const lines = [];
    let line = [];
    for (let i=0;i<tokens.length;i++) {
      const t = tokens[i];
      if (/^\d+\.$/.test(t)) {
        if (line.length) lines.push(line.join(' '));
        line = [t];
      } else if (/^(1-0|0-1|1\/2-1\/2|\*)$/.test(t)) {
        if (line.length) { lines.push(line.join(' ')); line = []; }
      } else {
        line.push(t);
      }
    }
    if (line.length) lines.push(line.join(' '));
    return lines.join('\n');
  }

  function formatUCIMoves(uciList){
    if (!uciList || !uciList.length) return '';
    let out = [];
    for (let i = 0, n = 1; i < uciList.length; i += 2, n++) {
      const w = uciList[i] || '';
      const b = uciList[i+1] || '';
      out.push(b ? (n + '. ' + w + ' ' + b) : (n + '. ' + w));
    }
    return out.join('\n');
  }

  function renderCaptured(byWhite, byBlack){
    capWhiteEl.textContent = '';
    capBlackEl.textContent = '';
    for (var i=0;i<byWhite.length;i++){
      var s1=document.createElement('span'); s1.textContent=byWhite[i]; capWhiteEl.appendChild(s1);
    }
    for (var j=0;j<byBlack.length;j++){
      var s2=document.createElement('span'); s2.textContent=byBlack[j]; capBlackEl.appendChild(s2);
    }
  }

  function capKey(id){ return 'tinychess:' + String(id||'') + ':captured:v1'; }

  // Prefill from storage to avoid blank on reload
  try{
    var saved = JSON.parse(localStorage.getItem(capKey(gameId)) || 'null');
    if (saved && saved.byWhite && saved.byBlack) renderCaptured(saved.byWhite, saved.byBlack);
  }catch(e){}

  // controls
  document.getElementById('reset').addEventListener('click', async ()=>{
    if(!gameId) return;
    await fetch('/reset/' + gameId, {method:'POST'});
    try { localStorage.removeItem(capKey(gameId)); } catch(e) {}
    renderCaptured([], []);
  });
  document.getElementById('copy').addEventListener('click', async ()=>{ try{ await navigator.clipboard.writeText(location.href); status('Link copied!'); setTimeout(()=>status(''),1200);}catch{ status('Copy failed', true); } });

  // live updates
  if (gameId){
    const es = new EventSource('/sse/' + gameId);
    es.onmessage = (ev)=>{
      const st = JSON.parse(ev.data);
      if (st.kind === 'emoji'){
        if (st.sender !== clientId) {
          showReaction(st.emoji); // only others see burst
        }
        return;
      }
      if (st.kind === 'state'){
        renderFEN(st.fen);
        turnEl.textContent = st.turn;
        pgnEl.textContent  = formatPGNLines(st.pgn || '');
        lanEl.textContent  = formatUCIMoves(st.uci || []);
        status(st.status || '');
        const caps = capturedFromFEN(st.fen);
        renderCaptured(caps.byWhite, caps.byBlack);
        try{ localStorage.setItem(capKey(gameId), JSON.stringify(caps)); }catch{}

        // Persist summary to recent list
        var resultFromPGN = (function(){
          var txt = (st.pgn || '').trim();
          var m = txt.match(/\b(1-0|0-1|1\/2-1\/2|\*)\b\s*$/);
          return m ? (m[1] === '*' ? null : m[1]) : null;
        })();
        var finishedNow = !!resultFromPGN || (!!st.status && /(1-0|0-1|1\/2-1\/2)/.test(st.status||''));
        setGameState(gameId, {
          moves: Array.isArray(st.uci) ? st.uci.length : 0,
          status: st.status || '',
          result: resultFromPGN || (function(){ var m = (st.status||'').match(/(1-0|0-1|1\/2-1\/2)/); return m ? m[1] : null; })(),
          finishedAt: finishedNow ? Date.now() : undefined,
          lastSeen: st.lastSeen
        });
      }
    };
    es.onerror = ()=>{ status('Disconnected. Reconnecting…', true); };
  }
})();
</script>
</body>
</html>`
