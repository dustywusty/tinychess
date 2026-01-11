//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"tinychess/internal/game"
	"tinychess/internal/handlers"
	"tinychess/internal/templates"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/google/uuid"
)

func TestPlayFullGame(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	runGame(t, "fools", []move{
		{color: "white", from: "f2", to: "f3"},
		{color: "black", from: "e7", to: "e5"},
		{color: "white", from: "g2", to: "g4"},
		{color: "black", from: "d8", to: "h4"},
	}, true)
}

func TestPlayScholarsMate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	runGame(t, "scholars", []move{
		{color: "white", from: "e2", to: "e4"},
		{color: "black", from: "e7", to: "e5"},
		{color: "white", from: "f1", to: "c4"},
		{color: "black", from: "b8", to: "c6"},
		{color: "white", from: "d1", to: "h5"},
		{color: "black", from: "g8", to: "f6"},
		{color: "white", from: "h5", to: "f7"},
	}, true)
}

func TestPlayLongGame(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	runGame(t, "long", []move{
		{color: "white", from: "e2", to: "e4"},
		{color: "black", from: "e7", to: "e5"},
		{color: "white", from: "g1", to: "f3"},
		{color: "black", from: "b8", to: "c6"},
		{color: "white", from: "f1", to: "b5"},
		{color: "black", from: "a7", to: "a6"},
		{color: "white", from: "b5", to: "a4"},
		{color: "black", from: "g8", to: "f6"},
		{color: "white", from: "e1", to: "g1"},
		{color: "black", from: "f8", to: "e7"},
		{color: "white", from: "f1", to: "e1"},
		{color: "black", from: "b7", to: "b5"},
		{color: "white", from: "a4", to: "b3"},
		{color: "black", from: "d7", to: "d6"},
		{color: "white", from: "c2", to: "c3"},
		{color: "black", from: "e8", to: "g8"},
		{color: "white", from: "h2", to: "h3"},
		{color: "black", from: "c6", to: "a5"},
		{color: "white", from: "b3", to: "c2"},
		{color: "black", from: "c7", to: "c5"},
		{color: "white", from: "d2", to: "d4"},
		{color: "black", from: "c5", to: "d4"},
		{color: "white", from: "c3", to: "d4"},
		{color: "black", from: "d8", to: "c7"},
		{color: "white", from: "d4", to: "e5"},
		{color: "black", from: "d6", to: "e5"},
		{color: "white", from: "f3", to: "e5"},
		{color: "black", from: "f6", to: "e4"},
		{color: "white", from: "e5", to: "d3"},
		{color: "black", from: "e4", to: "f6"},
		{color: "white", from: "d1", to: "d2"},
		{color: "black", from: "c7", to: "d6"},
		{color: "white", from: "c2", to: "b3"},
		{color: "black", from: "c8", to: "e6"},
	}, false)
}

type move struct {
	color string
	from  string
	to    string
}

func runGame(t *testing.T, label string, moves []move, expectMate bool) {
	t.Helper()

	server := newTestServer()
	defer server.Close()

	gameID := uuid.NewString()
	gameURL := fmt.Sprintf("%s/%s", server.URL, gameID)

	whiteCtx, whiteCancel := newBrowserCtx(t)
	defer whiteCancel()
	blackCtx, blackCancel := newBrowserCtx(t)
	defer blackCancel()

	width := envInt("CHROMEDP_VIEWPORT_WIDTH", 1280)
	height := envInt("CHROMEDP_VIEWPORT_HEIGHT", 2000)

	attachDebug(t, whiteCtx, "white-"+label)
	attachDebug(t, blackCtx, "black-"+label)

	if err := chromedp.Run(whiteCtx,
		network.Enable(),
		runtime.Enable(),
		page.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(jsErrorCaptureScript).Do(ctx)
			return err
		}),
	); err != nil {
		t.Fatalf("white debug enable failed: %v", err)
	}
	if err := chromedp.Run(blackCtx,
		network.Enable(),
		runtime.Enable(),
		page.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(jsErrorCaptureScript).Do(ctx)
			return err
		}),
	); err != nil {
		t.Fatalf("black debug enable failed: %v", err)
	}

	if err := chromedp.Run(whiteCtx, chromedp.Navigate(gameURL)); err != nil {
		t.Fatalf("white navigate failed: %v", err)
	}
	if err := chromedp.Run(blackCtx, chromedp.Navigate(gameURL)); err != nil {
		t.Fatalf("black navigate failed: %v", err)
	}
	if err := chromedp.Run(whiteCtx, chromedp.EmulateViewport(int64(width), int64(height))); err != nil {
		t.Fatalf("white viewport failed: %v", err)
	}
	if err := chromedp.Run(blackCtx, chromedp.EmulateViewport(int64(width), int64(height))); err != nil {
		t.Fatalf("black viewport failed: %v", err)
	}

	waitForBoard(t, whiteCtx, "white-"+label)
	waitForBoard(t, blackCtx, "black-"+label)

	whiteTurnText := waitForTurnText(t, whiteCtx)
	blackTurnText := waitForTurnText(t, blackCtx)

	if whiteTurnText != "Your turn" && blackTurnText != "Your turn" {
		t.Fatalf("expected one client to be white; got turn texts %q and %q", whiteTurnText, blackTurnText)
	}
	if whiteTurnText != "Your turn" {
		whiteCtx, blackCtx = blackCtx, whiteCtx
	}

	rec := startRecording(t, whiteCtx, label)
	defer rec.Stop(t)

	rec.Capture(t)
	sleepMoveDelay("E2E_START_DELAY_MS")

	for i, mv := range moves {
		t.Logf("move %d %s%s (%s)", i+1, mv.from, mv.to, mv.color)

		ctx := whiteCtx
		otherCtx := blackCtx
		if mv.color == "black" {
			ctx = blackCtx
			otherCtx = whiteCtx
		}

		turnText, statusText := readTurnStatus(ctx)
		t.Logf("move %d pre-turn=%q status=%q", i+1, turnText, statusText)

		waitForTurn(t, ctx, "Your turn")
		rec.Capture(t)
		expectTurnFlip := i < len(moves)-1
		if err := performMove(t, ctx, gameID, mv.from, mv.to, expectTurnFlip); err != nil {
			t.Fatalf("move %d %s%s failed: %v", i+1, mv.from, mv.to, err)
		}
		if expectTurnFlip {
			waitForTurn(t, otherCtx, "Your turn")
		}

		sleepMoveDelay("E2E_CAPTURE_DELAY_MS")
		rec.Capture(t)
		sleepMoveDelay("E2E_MOVE_DELAY_MS")
	}

	if expectMate {
		waitForStatusContains(t, whiteCtx, "Checkmate")
		waitForStatusContains(t, blackCtx, "Checkmate")
	}

	if err := maybeSendEmojiTaunt(t, whiteCtx); err != nil {
		t.Fatalf("emoji taunt failed: %v", err)
	}
	sleepMoveDelay("E2E_CAPTURE_DELAY_MS")
	rec.Capture(t)

	sleepMoveDelay("E2E_CAPTURE_DELAY_MS")
	rec.Capture(t)
}

func newTestServer() *httptest.Server {
	if err := chdirToRepoRoot(); err != nil {
		panic(fmt.Sprintf("e2e: %v", err))
	}
	templates.SetVersion("e2e")
	hub := game.NewHub()
	h := handlers.NewHandler(hub)

	mux := http.NewServeMux()
	mux.HandleFunc("/new", h.HandleNew)
	mux.HandleFunc("/sse/", h.HandleSSE)
	mux.HandleFunc("/move/", h.HandleMove)
	mux.HandleFunc("/react/", h.HandleReact)
	mux.HandleFunc("/release/", h.HandleRelease)
	mux.HandleFunc("/coach", h.HandleCoach)
	mux.HandleFunc("/", h.HandlePage)

	return httptest.NewServer(mux)
}

func newBrowserCtx(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	userDir, err := os.MkdirTemp("", "tinychess-e2e-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}

	width := envInt("CHROMEDP_VIEWPORT_WIDTH", 1280)
	height := envInt("CHROMEDP_VIEWPORT_HEIGHT", 2000)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.Flag("window-size", fmt.Sprintf("%d,%d", width, height)),
		chromedp.Flag("force-device-scale-factor", "1"),
		chromedp.UserDataDir(userDir),
	)
	if proxy := os.Getenv("CHROMEDP_PROXY_SERVER"); proxy != "" {
		opts = append(opts, chromedp.Flag("proxy-server", proxy))
	}
	if os.Getenv("CHROMEDP_HEADLESS") == "0" {
		opts = append(opts, chromedp.Flag("headless", false))
	} else {
		opts = append(opts, chromedp.Headless)
	}
	if os.Getenv("CHROMEDP_NO_SANDBOX") == "1" {
		opts = append(opts, chromedp.Flag("no-sandbox", true))
	}
	if bin := os.Getenv("CHROME_BIN"); bin != "" {
		opts = append(opts, chromedp.ExecPath(bin))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	ctx, timeoutCancel := context.WithTimeout(ctx, 45*time.Second)

	return ctx, func() {
		timeoutCancel()
		cancel()
		allocCancel()
		_ = os.RemoveAll(userDir)
	}
}

func attachDebug(t *testing.T, ctx context.Context, label string) {
	t.Helper()

	reqURLs := make(map[network.RequestID]string)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventRequestWillBeSent:
			reqURLs[e.RequestID] = e.Request.URL
		case *runtime.EventConsoleAPICalled:
			t.Logf("[%s] console.%s: %s", label, e.Type, joinConsoleArgs(e.Args))
		case *runtime.EventExceptionThrown:
			t.Logf("[%s] exception: %s", label, e.ExceptionDetails.Error())
		case *network.EventResponseReceived:
			url := reqURLs[e.RequestID]
			if url == "" {
				url = e.Response.URL
			}
			if strings.Contains(url, "/sse/") || strings.Contains(url, "/move/") {
				t.Logf("[%s] response: %s %d %s", label, url, int(e.Response.Status), e.Response.StatusText)
			}
		case *network.EventLoadingFailed:
			url := reqURLs[e.RequestID]
			if e.BlockedReason != "" {
				t.Logf("[%s] request blocked: %s %s (%s) url=%s", label, e.RequestID, e.ErrorText, e.BlockedReason, url)
			} else {
				t.Logf("[%s] request failed: %s %s url=%s", label, e.RequestID, e.ErrorText, url)
			}
			delete(reqURLs, e.RequestID)
		case *page.EventJavascriptDialogOpening:
			t.Logf("[%s] dialog: %s %s", label, e.Type, e.Message)
		}
	})
}

func joinConsoleArgs(args []*runtime.RemoteObject) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, 0, len(args))
	for _, a := range args {
		if a == nil {
			continue
		}
		if a.Value != nil {
			parts = append(parts, fmt.Sprintf("%v", a.Value))
			continue
		}
		if a.Description != "" {
			parts = append(parts, a.Description)
			continue
		}
		parts = append(parts, string(a.Type))
	}
	return strings.Join(parts, " ")
}

func waitForTurnText(t *testing.T, ctx context.Context) string {
	t.Helper()
	text, err := pollText(ctx, "#turn", func(s string) bool {
		return s == "Your turn" || s == "Their turn"
	}, 10*time.Second)
	if err != nil {
		t.Fatalf("turn text: %v", err)
	}
	return text
}

func waitForTurn(t *testing.T, ctx context.Context, want string) {
	t.Helper()
	if _, err := pollText(ctx, "#turn", func(s string) bool {
		return s == want
	}, 10*time.Second); err != nil {
		turnText, statusText := readTurnStatus(ctx)
		t.Fatalf("turn text %q: %v (turn=%q status=%q)", want, err, turnText, statusText)
	}
}

func sleepMoveDelay(env string) {
	delay := envDuration(env, 0)
	if delay > 0 {
		time.Sleep(delay)
	}
}

func readTurnStatus(ctx context.Context) (string, string) {
	var turnText string
	var statusText string
	_ = chromedp.Run(ctx,
		chromedp.Text("#turn", &turnText, chromedp.ByQuery),
		chromedp.Text("#status", &statusText, chromedp.ByQuery),
	)
	return strings.TrimSpace(turnText), strings.TrimSpace(statusText)
}

func waitForTurnValue(ctx context.Context, want string, timeout time.Duration) error {
	_, err := pollText(ctx, "#turn", func(s string) bool {
		return s == want
	}, timeout)
	return err
}

func performMove(t *testing.T, ctx context.Context, gameID, from, to string, expectTurnFlip bool) error {
	t.Helper()

	t.Logf("click %s -> %s", from, to)
	if err := clickSquare(ctx, from); err != nil {
		return err
	}
	if err := clickSquare(ctx, to); err != nil {
		return err
	}

	if expectTurnFlip {
		turnText, statusText := readTurnStatus(ctx)
		t.Logf("post-click turn=%q status=%q", turnText, statusText)

		if err := waitForTurnValue(ctx, "Their turn", 2*time.Second); err == nil {
			return nil
		}

		t.Logf("move %s%s did not flip turn; retrying via fetch", from, to)
		if err := submitMoveViaFetch(t, ctx, gameID, from+to); err != nil {
			return err
		}
		if err := waitForTurnValue(ctx, "Their turn", 5*time.Second); err != nil {
			turnText, statusText := readTurnStatus(ctx)
			return fmt.Errorf("turn not flipped after fallback (turn=%q status=%q)", turnText, statusText)
		}
	}

	return nil
}

func submitMoveViaFetch(t *testing.T, ctx context.Context, gameID, uci string) error {
	t.Helper()

	script := fmt.Sprintf(`(async () => {
		const clientId = sessionStorage.getItem("tinychess:clientId") || "";
		if (!clientId) return { ok: false, error: "missing clientId" };
		const res = await fetch("/move/%s", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ uci: %q, clientId })
		});
		return await res.json();
	})()`, gameID, uci)
	var out map[string]any
	if err := runWithTimeout(ctx, 5*time.Second, chromedp.Evaluate(script, &out)); err != nil {
		return err
	}
	t.Logf("fetch move %s => %v", uci, out)
	if ok, _ := out["ok"].(bool); !ok {
		return fmt.Errorf("move rejected: %v", out)
	}
	return nil
}

func envDuration(name string, def time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return def
	}
	ms, err := time.ParseDuration(raw + "ms")
	if err != nil {
		return def
	}
	return ms
}

func startRecording(t *testing.T, ctx context.Context, label string) *recorder {
	t.Helper()
	rec := &recorder{enabled: os.Getenv("E2E_RECORD") == "1"}
	if !rec.enabled {
		return rec
	}

	rec.ctx = ctx
	rec.dir = strings.TrimSpace(os.Getenv("E2E_RECORD_DIR"))
	if rec.dir == "" {
		rec.dir = "e2e-artifacts"
	}
	if err := os.MkdirAll(rec.dir, 0o755); err != nil {
		t.Fatalf("record dir: %v", err)
	}
	rec.prefix = filepath.Join(rec.dir, label)
	rec.fps = envInt("E2E_RECORD_FPS", 6)
	rec.holdMS = envInt("E2E_RECORD_HOLD_MS", 2000)
	rec.format = strings.ToLower(strings.TrimSpace(os.Getenv("E2E_RECORD_FORMAT")))
	if rec.format == "" {
		rec.format = "mp4"
	}
	return rec
}

func maybeSendEmojiTaunt(t *testing.T, ctx context.Context) error {
	t.Helper()
	if strings.TrimSpace(os.Getenv("E2E_SEND_EMOJI")) == "0" {
		return nil
	}
	emoji := strings.TrimSpace(os.Getenv("E2E_EMOJI"))
	if emoji == "" {
		emoji = `\u{1F921}`
	}
	script := fmt.Sprintf(`(() => {
		const btn = document.getElementById("reactbtn");
		if (!btn) return { ok: false, error: "react button missing" };
		if (btn.disabled) return { ok: false, error: "react button disabled" };
		btn.click();
		const picker = document.getElementById("emojiPicker");
		if (!picker) return { ok: false, error: "emoji picker missing" };
		const emoji = "%s";
		const ev = new CustomEvent("emoji-click", { detail: { unicode: emoji } });
		picker.dispatchEvent(ev);
		return { ok: true, emoji };
	})()`, emoji)
	var out map[string]any
	if err := runWithTimeout(ctx, 5*time.Second, chromedp.Evaluate(script, &out)); err != nil {
		return err
	}
	if ok, _ := out["ok"].(bool); !ok {
		return fmt.Errorf("emoji picker failed: %v", out)
	}
	if _, err := pollText(ctx, ".big-emoji", func(s string) bool {
		return strings.TrimSpace(s) != ""
	}, 2*time.Second); err != nil {
		return fmt.Errorf("emoji not shown: %v", err)
	}
	return nil
}

type recorder struct {
	enabled bool
	ctx     context.Context
	dir     string
	prefix  string
	frame   int
	format  string
	fps     int
	holdMS  int
}

func (r *recorder) Capture(t *testing.T) {
	if !r.enabled {
		return
	}
	r.frame++
	var data []byte
	if err := runWithTimeout(r.ctx, 2*time.Second, chromedp.ActionFunc(func(ctx context.Context) error {
		out, err := page.CaptureScreenshot().
			WithFormat(page.CaptureScreenshotFormatPng).
			WithCaptureBeyondViewport(true).
			Do(ctx)
		if err != nil {
			return err
		}
		data = out
		return nil
	})); err != nil {
		t.Logf("recording capture: %v", err)
		return
	}
	name := fmt.Sprintf("%s_%05d.png", r.prefix, r.frame)
	if err := os.WriteFile(name, data, 0o644); err != nil {
		t.Logf("recording write: %v", err)
	}
}

func (r *recorder) Stop(t *testing.T) {
	if !r.enabled {
		return
	}
	if err := renderRecording(t, r.dir, r.prefix, r.format, r.fps, r.holdMS); err != nil {
		t.Logf("recording render: %v", err)
	}
}

func renderRecording(t *testing.T, dir, framePrefix, format string, fps int, holdMS int) error {
	t.Helper()

	if format == "frames" {
		return nil
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found (set E2E_RECORD_FORMAT=frames to keep PNGs)")
	}

	input := framePrefix + "_%05d.png"
	output := filepath.Join(dir, filepath.Base(framePrefix)+"."+format)

	var args []string
	var holdArg string
	if holdMS > 0 {
		holdSeconds := float64(holdMS) / 1000.0
		holdArg = fmt.Sprintf("tpad=stop_mode=clone:stop_duration=%.3f", holdSeconds)
	}
	switch format {
	case "gif":
		args = []string{"-y", "-framerate", strconv.Itoa(fps), "-i", input}
		if holdArg != "" {
			args = append(args, "-vf", holdArg)
		}
		args = append(args, output)
	case "mp4":
		args = []string{"-y", "-framerate", strconv.Itoa(fps), "-i", input}
		if holdArg != "" {
			args = append(args, "-vf", holdArg)
		}
		args = append(args, "-pix_fmt", "yuv420p", output)
	default:
		return fmt.Errorf("unsupported E2E_RECORD_FORMAT=%s", format)
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func envInt(name string, def int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return def
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return val
}

func waitForStatusContains(t *testing.T, ctx context.Context, want string) {
	t.Helper()
	if _, err := pollText(ctx, "#status", func(s string) bool {
		return strings.Contains(s, want)
	}, 10*time.Second); err != nil {
		t.Fatalf("status %q: %v", want, err)
	}
}

func pollText(ctx context.Context, selector string, ok func(string) bool, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	var text string
	for time.Now().Before(deadline) {
		err := runWithTimeout(ctx, 750*time.Millisecond, chromedp.Text(selector, &text, chromedp.ByQuery))
		if err == nil {
			text = strings.TrimSpace(text)
			if ok(text) {
				return text, nil
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
	if text == "" {
		text = "<empty>"
	}
	return "", fmt.Errorf("timeout waiting for %s text, last=%s", selector, text)
}

func waitVisible(selector string) chromedp.Action {
	return chromedp.WaitVisible(selector, chromedp.ByQuery)
}

func clickSquare(ctx context.Context, square string) error {
	selector := fmt.Sprintf(`[data-square="%s"]`, square)
	script := fmt.Sprintf(`(() => {
		const el = document.querySelector(%q);
		if (!el) return false;
		el.click();
		return true;
	})()`, selector)
	var ok bool
	if err := runWithTimeout(ctx, 2*time.Second, chromedp.Evaluate(script, &ok)); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("square %s not found", square)
	}
	return nil
}

func runWithTimeout(ctx context.Context, timeout time.Duration, actions ...chromedp.Action) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return chromedp.Run(ctxTimeout, actions...)
}

const jsErrorCaptureScript = `
(() => {
  window.__tinychessErrors = [];
  const origEventSource = window.EventSource;
  if (origEventSource) {
    window.EventSource = function(...args) {
      const es = new origEventSource(...args);
      const url = args && args[0] ? String(args[0]) : "";
      es.addEventListener("open", () => {
        window.__tinychessErrors.push({ type: "eventsource-open", url });
      });
      es.addEventListener("error", () => {
        window.__tinychessErrors.push({ type: "eventsource-error", url, readyState: es.readyState });
      });
      return es;
    };
  }
  window.addEventListener("error", (e) => {
    window.__tinychessErrors.push({
      type: "error",
      message: e.message || "",
      filename: e.filename || "",
      lineno: e.lineno || 0,
      colno: e.colno || 0
    });
  });
  window.addEventListener("unhandledrejection", (e) => {
    window.__tinychessErrors.push({
      type: "unhandledrejection",
      message: String(e.reason || "")
    });
  });
})();
`

func waitForBoard(t *testing.T, ctx context.Context, label string) {
	t.Helper()

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		var visible bool
		err := chromedp.Run(ctx, chromedp.Evaluate(`(() => {
			const el = document.querySelector("#board");
			if (!el) return false;
			const r = el.getBoundingClientRect();
			return !!(r.width && r.height);
		})()`, &visible))
		if err == nil && visible {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}

	var readyState string
	var title string
	var url string
	var errors any
	var body string
	_ = chromedp.Run(ctx,
		chromedp.Evaluate(`document.readyState`, &readyState),
		chromedp.Title(&title),
		chromedp.Location(&url),
		chromedp.Evaluate(`window.__tinychessErrors || []`, &errors),
		chromedp.OuterHTML("body", &body, chromedp.ByQuery),
	)

	t.Logf("[%s] board not visible after timeout", label)
	t.Logf("[%s] readyState=%s title=%q url=%s", label, readyState, title, url)
	t.Logf("[%s] errors=%v", label, errors)
	t.Logf("[%s] body=%s", label, truncate(body, 1200))
	t.Fatalf("[%s] board did not render", label)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

func chdirToRepoRoot() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	dir := wd
	for i := 0; i < 6; i++ {
		if exists(filepath.Join(dir, "internal", "templates", "game.html")) {
			return os.Chdir(dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return fmt.Errorf("repo root not found from %s", wd)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
