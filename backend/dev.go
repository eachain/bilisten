//go:build ignore

package main

import (
	"net/http"

	"bilisten/backend"
)

func main() {
	backend.ServerOrigin = "http://127.0.0.1:65511"
	const DataDir = "tmp/data"
	const CacheDir = "tmp/cache"
	backend.InitLogger(CacheDir)
	backend.InitImageCacheManager(DataDir, CacheDir)
	backend.InitAudioCacheManager(DataDir, CacheDir)
	backend.InitStorageManager(DataDir)

	panic(
		http.ListenAndServe(
			"127.0.0.1:65511",
			backend.NewRouter(http.NotFoundHandler()),
		),
	)
}
