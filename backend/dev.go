//go:build ignore

package main

import (
	"log"
	"net/http"

	"bilisten/backend"
)

func main() {
	backend.ServerOrigin = "http://127.0.0.1:65511"
	const DataDir = "tmp/data"
	const CacheDir = "tmp/cache"
	backend.InitImageCacheManager(DataDir, CacheDir)
	backend.InitAudioCacheManager(DataDir, CacheDir)
	backend.InitStorageManager(DataDir)

	// fp, _ := os.OpenFile("tmp/log.txt", os.O_TRUNC|os.O_WRONLY, 0644)
	// log.SetOutput(fp)
	log.SetFlags(log.Flags() | log.Lmicroseconds | log.Lshortfile)

	panic(
		http.ListenAndServe(
			"127.0.0.1:65511",
			backend.NewRouter(http.NotFoundHandler()),
		),
	)
}
