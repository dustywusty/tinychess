package frontend

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestStaticPages(t *testing.T) {
	const dir = "../../frontend"
	handler := Handler(dir)
	for _, tc := range []struct{ path, file string }{
		{"/", "index.html"},
		{"/shared-game", "game.html"},
		{"/another-game", "game.html"},
		{"/new/", "new/index.html"},
		{"/site.js", "site.js"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			r := httptest.NewRecorder()
			handler.ServeHTTP(r, httptest.NewRequest("GET", tc.path, nil))
			want, err := os.ReadFile(dir + "/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			if r.Code != 200 || r.Body.String() != string(want) {
				t.Fatalf("%s did not serve the unchanged static file: %d", tc.path, r.Code)
			}
			if strings.Contains(r.Body.String(), "{{") {
				t.Fatal("static file contains a server template placeholder")
			}
		})
	}
}
