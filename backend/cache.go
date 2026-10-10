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

	var cache *AudioCacheInfo
	for _, q := range SoundQuality {
		cache = globalAudioCacheManager.GetAudioInfo(bvid, cid, q)
		if cache != nil {
			break
		}
	}

	var qs []string
	if cache != nil {
		qs = cache.Qualities
	} else {
		infos, status := getAudioInfos(tx, bvid, cid)
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		for _, info := range infos {
			qs = append(qs, SoundQuality[info.Id])
		}
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
