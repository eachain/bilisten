package backend

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func queryAudioQuality(w http.ResponseWriter, r *http.Request) {
	tx := NewTx()
	query := r.URL.Query()
	bvid := query.Get("bvid")
	cid := query.Get("cid")
	infos, status := getAudioInfos(tx, bvid, cid)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	qs := make([]string, 0, len(infos))
	for _, info := range infos {
		qs = append(qs, SoundQuality[info.Id])
	}
	json.NewEncoder(w).Encode(qs)
}

func RouteCacheAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/cache/")
	switch path {
	case "audio/q":
		queryAudioQuality(w, r)

	case "audio/size":
		size := globalAudioCacheManager.GetUsedSize()
		json.NewEncoder(w).Encode(strconv.FormatInt(size, 10))

	case "audio/clear":
		ClearAudioCache()

	case "image/size":
		size := globalImageCacheManager.GetUsedSize()
		json.NewEncoder(w).Encode(strconv.FormatInt(size, 10))

	case "image/clear":
		ClearImageCache()

	default:
		http.NotFound(w, r)
	}
}
