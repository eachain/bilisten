package backend

import (
	"net/http"
	"strings"
)

// set them when dev
var ServerOrigin = ""

func AccessCors(header http.Header) {
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Credentials", "true")
	header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	header.Set("Access-Control-Allow-Headers", "Content-Length, Content-Type, Origin, Authorization, X-Requested-With")
	header.Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Access-Control-Allow-Methods")
	header.Set("Access-Control-Max-Age", "86400")
}

func NewRouter(notFound http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		AccessCors(w.Header())
		if r.Method == http.MethodOptions {
			return
		}

		switch {
		case strings.HasPrefix(r.URL.Path, "/api"):
			ProxyAPI(w, r)

		case strings.HasPrefix(r.URL.Path, "/image"):
			ProxyImage(w, r)

		case strings.HasPrefix(r.URL.Path, "/mpd/"):
			ProxyMPD(w, r)

		case strings.HasPrefix(r.URL.Path, "/audio/"):
			ProxyAudio(w, r)

		case strings.HasPrefix(r.URL.Path, "/cache/"):
			RouteCacheAPI(w, r)

		case strings.HasPrefix(r.URL.Path, "/store/"):
			RouteStoreAPI(w, r)
		default:
			notFound.ServeHTTP(w, r)
		}
	}
}
