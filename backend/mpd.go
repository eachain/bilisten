package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/eachain/lru"
)

type AudioInfo struct {
	Id          int         `json:"id"` // SoundQuality
	LocalURL    string      `json:"-"`
	Duration    int64       `json:"-"`
	BaseURL     string      `json:"base_url"`
	BackupURL   []string    `json:"backup_url"`
	Bandwidth   int         `json:"bandwidth"`
	MimeType    string      `json:"mime_type"` // "audio/mp4"
	Codecs      string      `json:"codecs"`    // "mp4a.40.2"
	SegmentBase SegmentBase `json:"SegmentBase"`
	Qualities   []string    `json:"qs,omitempty"`
}

func ContentTypeOf(mimeType string, codecs string) string {
	if mimeType != "" && codecs != "" {
		return fmt.Sprintf("%v; codecs=%q", mimeType, codecs)
	}
	return ""
}

type AudioInfoCache struct {
	Infos []*AudioInfo
	Time  time.Time
}

var (
	audioRealInfoCache = lru.New[string, *AudioInfoCache](1000)
	audioCacheMutex    = new(sync.RWMutex)
)

func evictExpired() {
	audioCacheMutex.Lock()
	defer audioCacheMutex.Unlock()

	now := time.Now()
	for key, cache := range audioRealInfoCache.All() {
		if now.Sub(cache.Time) > time.Hour {
			audioRealInfoCache.Remove(key)
		}
	}
}

func audioCacheKey(bvid, cid string) string {
	return fmt.Sprintf("%v-%v", bvid, cid)
}

func GetAudioInfoFromCache(bvid, cid string) []*AudioInfo {
	evictExpired()

	audioCacheMutex.RLock()
	defer audioCacheMutex.RUnlock()

	cache, _ := audioRealInfoCache.Get(audioCacheKey(bvid, cid))
	if cache == nil {
		return nil
	}
	return cache.Infos
}

func SetAudioInfoToCache(bvid, cid string, infos []*AudioInfo) {
	evictExpired()

	audioCacheMutex.Lock()
	defer audioCacheMutex.Unlock()

	audioRealInfoCache.Set(audioCacheKey(bvid, cid), &AudioInfoCache{
		Infos: infos,
		Time:  time.Now(),
	})
}

type biliResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type dolbyInfo struct {
	Audio []*AudioInfo `json:"audio"`
}

type flacInfo struct {
	Audio *AudioInfo `json:"audio"`
}

type playResp struct {
	Dash struct {
		Duration int64        `json:"duration"`
		Audio    []*AudioInfo `json:"audio"`
		Dolby    *dolbyInfo   `json:"dolby"`
		Flac     *flacInfo    `json:"flac"`
	} `json:"dash"`
}

var SoundQuality = map[int]string{
	30216: "low",   // 64kbps, size: 2,856,834 = 2MB
	30232: "mid",   // 132kbps, size: 4,323,852 = 4MB
	30280: "high",  // 192kbps, size: 8,592,530 = 8MB
	30250: "dolby", // 1025kbps, size: 44,522,869 = 44MB
	30251: "flac",  // 1584kbps, size: 68,874,306 = 68MB
}

var QualityOf = map[string]int{
	"64kbps":  30216,
	"132kbps": 30232,
	"192kbps": 30280,

	"low":   30216,
	"mid":   30232,
	"high":  30280,
	"dolby": 30250,
	"flac":  30251,
}

func getLocalURL(bvid, cid, q string) string {
	return fmt.Sprintf("%v/audio/%v/%v/%v.m4s", ServerOrigin, bvid, cid, q)
}

func getAudioPlayURL(tx string, bvid, cid string) ([]*AudioInfo, int) {
	query := make(url.Values)
	query.Add("bvid", bvid)
	query.Add("cid", cid)
	query.Add("fnval", "4048") // 16: dash, 256: 杜比音频
	query.Add("fourk", "1")
	query.Add("qn", "0")
	query.Add("from_client", "BROWSER")
	query.Add("is_main_page", "true")
	url := `https://api.bilibili.com/x/player/playurl?` + query.Encode()

	start := time.Now()
	log := NewLogger("getAudioPlayURL").Tx(tx)
	resp := DoRequestAPI(log, &Request{
		Tx:  tx,
		URL: url,
	})
	log.Cost(time.Since(start))
	if resp.StatusCode != http.StatusOK {
		log.Error("http status %v", resp.StatusCode)
		return nil, resp.StatusCode
	}

	br := new(biliResponse)
	err := json.Unmarshal(resp.Body, br)
	if err != nil {
		log.Error("json decode response: %v", err)
		return nil, http.StatusInternalServerError
	}
	if br.Code != 0 {
		log.Error("response code: %v %v", br.Code, br.Message)
		return nil, http.StatusBadRequest
	}

	pr := new(playResp)
	err = json.Unmarshal(br.Data, pr)
	if err != nil {
		log.Error("json decode response playurl: %v", err)
		return nil, http.StatusInternalServerError
	}
	audios := pr.Dash.Audio
	var qs []string
	if pr.Dash.Flac != nil && pr.Dash.Flac.Audio != nil {
		audios = append(audios, pr.Dash.Flac.Audio)
	}
	if pr.Dash.Dolby != nil && len(pr.Dash.Dolby.Audio) > 0 {
		audios = append(audios, pr.Dash.Dolby.Audio...)
	}
	if len(audios) == 0 {
		log.Error("no audios found")
		return nil, http.StatusInternalServerError
	}
	for _, info := range audios {
		info.Duration = pr.Dash.Duration
		q := SoundQuality[info.Id]
		if q == "" {
			q = strconv.Itoa(info.Id)
		}
		qs = append(qs, q)
		info.LocalURL = getLocalURL(bvid, cid, q)
	}
	for _, info := range audios {
		info.Qualities = qs
	}
	SetAudioInfoToCache(bvid, cid, audios)

	log.Info("done")

	return audios, http.StatusOK
}

func getAudioInfos(tx string, bvid, cid string) (infos []*AudioInfo, status int) {
	status = http.StatusOK
	infos = GetAudioInfoFromCache(bvid, cid)
	if len(infos) == 0 {
		infos, status = getAudioPlayURL(tx, bvid, cid)
	}
	return
}

func getAudioInfo(tx string, bvid, cid string, qs []string) (info *AudioInfo, status int) {
	var infos []*AudioInfo
	infos, status = getAudioInfos(tx, bvid, cid)
	if status != http.StatusOK {
		return
	}

	for _, q := range qs {
		id := QualityOf[q]
		for _, a := range infos {
			if a.Id == id {
				info = a
				return
			}
		}
	}

	if len(infos) > 0 {
		info = infos[0]
		return
	}

	status = http.StatusInternalServerError
	return
}

const MPDTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011"
     type="static"
     minBufferTime="PT1S"
     mediaPresentationDuration="PT{{.Duration}}S"
     profiles="urn:mpeg:dash:profile:isoff-on-demand:2011">
  <Period>
    <AdaptationSet contentType="audio" mimeType="{{.MimeType}}">
      <Representation
          id="{{.Id}}"
          bandwidth="{{.Bandwidth}}"
          codecs="{{.Codecs}}">
        <BaseURL>{{.LocalURL}}</BaseURL>
        <SegmentBase indexRange="{{.SegmentBase.IndexRange}}">
          <Initialization range="{{.SegmentBase.Initialization}}"/>
        </SegmentBase>
      </Representation>
    </AdaptationSet>
  </Period>
</MPD>`

var mpdTemp *template.Template

func init() {
	temp, err := template.New("mpd").Parse(MPDTemplate)
	if err != nil {
		panic(err)
	}
	mpdTemp = temp
}

func parseMpdBvidCidQuality(u *url.URL) (bvid, cid string, qs []string, ok bool) {
	fields := strings.Split(u.Path, "/")
	if len(fields) < 4 {
		return
	}
	// "/mpd/bvid/cid.mpd?q=flac,dolby,high,mid,low"
	bvid = fields[2]
	cid = strings.TrimSuffix(fields[3], ".mpd")
	qs = strings.Split(u.Query().Get("q"), ",")
	if len(qs) == 1 && qs[0] == "" {
		qs[0] = "mid"
	}
	ok = true
	return
}

func ProxyMPD(w http.ResponseWriter, r *http.Request) {
	tx := NewTx()
	bvid, cid, qs, ok := parseMpdBvidCidQuality(r.URL)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var cache *AudioCacheInfo
	for i, q := range qs {
		c := globalAudioCacheManager.GetAudioInfo(bvid, cid, q)
		if c == nil {
			continue
		}

		exists := false
		for _, pq := range qs[:i] {
			if slices.Contains(c.Qualities, pq) {
				exists = true
				break
			}
		}
		if !exists {
			cache = c
		}
		break
	}

	var info *AudioInfo
	if cache != nil {
		info = &AudioInfo{
			Id:          cache.Id,
			LocalURL:    getLocalURL(bvid, cid, SoundQuality[cache.Id]),
			Duration:    cache.Duration,
			BaseURL:     "",
			BackupURL:   nil,
			Bandwidth:   cache.Bandwidth,
			MimeType:    cache.MimeType,
			Codecs:      cache.Codecs,
			SegmentBase: cache.SegmentBase,
			Qualities:   cache.Qualities,
		}
	} else {
		var status int
		info, status = getAudioInfo(tx, bvid, cid, qs)
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}

		query := r.URL.Query()
		cache = &AudioCacheInfo{
			Bvid:        bvid,
			Cid:         cid,
			Title:       query.Get("title"),
			Pic:         query.Get("pic"),
			Author:      query.Get("author"),
			Mid:         query.Get("mid"),
			Id:          info.Id,
			Duration:    info.Duration,
			Size:        0,
			Bandwidth:   info.Bandwidth,
			MimeType:    info.MimeType,
			Codecs:      info.Codecs,
			SegmentBase: info.SegmentBase,
			Qualities:   info.Qualities,
		}
		globalAudioCacheManager.SetAudioInfo(tx, cache)
	}

	var buf bytes.Buffer
	err := mpdTemp.Execute(&buf, info)
	if err != nil {
		NewLogger("ProxyMPD").Tx(tx).Error("mpd template execute: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/dash+xml")
	buf.WriteTo(w)
}
