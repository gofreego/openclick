package http_server

import (
	"net/http"
	"strings"
)

const PREFIX = "/openclick/"

// UI Handler with SPA Fallback
func GetUIHandler(uifs http.FileSystem, indexHTML []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.URL.Path, "assets") {
			http.StripPrefix(PREFIX, http.FileServer(uifs)).ServeHTTP(w, r)
			return
		}
		// Fallback to index.html for SPA routes
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(indexHTML)
	})
}
