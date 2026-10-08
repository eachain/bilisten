package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eachain/lru"
)

type searchManager struct {
	file string
	lru  *lru.MutexLRU[string, bool]
}

func (sm *searchManager) load() {
	content, err := os.ReadFile(sm.file)
	if err != nil {
		return
	}
	var keywords []string
	err = json.Unmarshal(content, &keywords)
	if err != nil {
		return
	}
	for _, keyword := range keywords {
		sm.lru.Set(keyword, true)
	}
}

func (sm *searchManager) save() {
	var keywords []string
	for keyword := range sm.lru.Backward() {
		keywords = append(keywords, keyword)
	}

	content, err := json.Marshal(keywords)
	if err != nil {
		return
	}

	os.WriteFile(sm.file, content, 0644)
}

func (sm *searchManager) all() []string {
	keywords := []string{}
	for keyword := range sm.lru.All() {
		keywords = append(keywords, keyword)
	}
	return keywords
}

func (sm *searchManager) clear() {
	sm.lru.Clear()
	sm.save()
}

func (sm *searchManager) set(keyword string) {
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		sm.lru.Set(keyword, true)
		sm.save()
	}
}

func (sm *searchManager) remove(keyword string) {
	_, ok := sm.lru.Remove(keyword)
	if ok {
		sm.save()
	}
}

func (sm *searchManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(sm.all())

	case http.MethodPost:
		sm.set(r.URL.Query().Get("keyword"))
		json.NewEncoder(w).Encode(sm.all())

	case http.MethodDelete:
		query := r.URL.Query()
		if query.Get("clear") == "true" {
			sm.clear()
		} else {
			sm.remove(query.Get("keyword"))
		}
		json.NewEncoder(w).Encode(sm.all())
	}
}

type playlistManager struct {
	file string
	list json.RawMessage
}

func (pm *playlistManager) load() {
	content, err := os.ReadFile(pm.file)
	if err != nil {
		return
	}
	if json.Valid(content) {
		pm.list = content
	} else {
		pm.list = json.RawMessage{'[', ']'}
	}
}

func (pm *playlistManager) update(list json.RawMessage) {
	if json.Valid(list) {
		pm.list = list
		os.WriteFile(pm.file, list, 0644)
	}
}

func (pm *playlistManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		w.Write(pm.list)

	case http.MethodPost:
		list, _ := io.ReadAll(r.Body)
		pm.update(list)
	}
}

type playstateManager struct {
	file  string
	Mode  string `json:"mode"`
	Index string `json:"index"`
}

func (pm *playstateManager) load() {
	content, err := os.ReadFile(pm.file)
	if err != nil {
		return
	}
	json.Unmarshal(content, pm)
}

func (pm *playstateManager) update(mode, index string) {
	var changed bool
	if mode != "" && mode != pm.Mode {
		pm.Mode = mode
		changed = true
	}
	if index != "" && index != pm.Index {
		pm.Index = index
		changed = true
	}
	if changed {
		content, err := json.Marshal(pm)
		if err != nil {
			return
		}
		os.WriteFile(pm.file, content, 0644)
	}
}

func (pm *playstateManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pm)

	case http.MethodPost:
		query := r.URL.Query()
		pm.update(query.Get("mode"), query.Get("index"))
	}
}

type settingsManager struct {
	file    string
	Buffer  int64  `json:"buffer"`
	Audio   int64  `json:"audio"`
	Image   int64  `json:"image"`
	Quality string `json:"quality"` // "mid,high,low"
}

func (sm *settingsManager) load() {
	content, err := os.ReadFile(sm.file)
	if err != nil {
		sm.Buffer = 30
		sm.Audio = 1000
		sm.Image = 10
		return
	}
	tmp := new(settingsManager)
	err = json.Unmarshal(content, tmp)
	if err != nil {
		return
	}
	sm.Buffer = tmp.Buffer
	if sm.Buffer <= 0 {
		sm.Buffer = 1
	}
	sm.Audio = tmp.Audio
	sm.Image = tmp.Image
	sm.Quality = tmp.Quality
}

func (sm *settingsManager) updateField(addr *int64, value string, minVal int64) bool {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	if val < minVal {
		return false
	}
	if *addr == val {
		return false
	}
	*addr = val
	return true
}

func (sm *settingsManager) update(key, value string) (result any) {
	updated := false
	switch key {
	case "audio":
		updated = sm.updateField(&sm.Audio, value, 0)
		if updated {
			SetAudioCacheSizeLimit(sm.Audio * 1024 * 1024)
		}
		result = sm.Audio

	case "image":
		updated = sm.updateField(&sm.Image, value, 0)
		if updated {
			SetImageCacheSizeLimit(sm.Image * 1024 * 1024)
		}
		result = sm.Image

	case "buffer":
		updated = sm.updateField(&sm.Buffer, value, 1)
		result = sm.Buffer

	case "quality":
		sm.Quality = value
		updated = true
		result = sm.Quality
	}

	if updated {
		content, err := json.Marshal(sm)
		if err != nil {
			return
		}
		os.WriteFile(sm.file, content, 0644)
	}

	return
}

func (sm *settingsManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var result any
	switch r.Method {
	case http.MethodGet:
		result = sm

	case http.MethodPost:
		query := r.URL.Query()
		result = sm.update(query.Get("key"), query.Get("value"))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

type cookiesManager struct {
	file    string
	mut     sync.RWMutex
	cookies []*http.Cookie
	expire  time.Time
	id      uint32
	wait    map[uint32]chan struct{}

	Cookies      []string `json:"cookies"`
	RefreshToken string   `json:"refresh_token"`
}

func (cm *cookiesManager) load() {
	cm.wait = make(map[uint32]chan struct{})

	content, err := os.ReadFile(cm.file)
	if err != nil {
		return
	}
	err = json.Unmarshal(content, cm)
	if err != nil {
		return
	}

	now := time.Now()
	cm.expire = now.Add(time.Hour * 24 * 365 * 10)
	for _, line := range cm.Cookies {
		cookie, err := http.ParseSetCookie(line)
		if err != nil || cookie == nil {
			continue
		}
		if cookie.Expires.After(now) {
			cm.cookies = append(cm.cookies, cookie)
			if cookie.Expires.Before(cm.expire) {
				cm.expire = cookie.Expires
			}
		}
	}
}

func (cm *cookiesManager) clear() {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	defer cm.onStatusChanged()

	if len(cm.cookies) == 0 {
		return
	}

	cm.cookies = nil
	cm.expire = time.Now()
	cm.Cookies = []string{}
	cm.RefreshToken = ""
	os.RemoveAll(cm.file)
}

func (cm *cookiesManager) clearExpiredCookies() {
	cm.mut.Lock()
	defer cm.mut.Unlock()
	if len(cm.cookies) == 0 {
		return
	}
	now := time.Now()
	if now.Before(cm.expire) {
		return
	}

	cm.cookies = nil
	cm.Cookies = []string{}
	cm.RefreshToken = ""
	os.RemoveAll(cm.file)

	cm.onStatusChanged()
}

func AddLoginedCookies(r *http.Request) {
	globalStorageManager.cookies.AddCookies(r)
}

func (cm *cookiesManager) AddCookies(r *http.Request) {
	cm.clearExpiredCookies()
	for _, c := range cm.cookies {
		r.AddCookie(c)
	}
}

func (cm *cookiesManager) SetCookies(cookies []*http.Cookie, refreshToken string) {
	if len(cookies) == 0 && refreshToken == "" {
		return
	}

	cm.mut.Lock()
	defer cm.mut.Unlock()

	defer cm.onStatusChanged()

	cm.cookies = cookies
	cm.RefreshToken = refreshToken
	now := time.Now()
	cm.expire = now.Add(time.Hour * 24 * 365 * 10)
	cm.Cookies = make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		cm.Cookies = append(cm.Cookies, cookie.String())
		if cookie.Expires.Before(cm.expire) {
			cm.expire = cookie.Expires
		}
	}

	content, err := json.Marshal(cm)
	if err != nil {
		return
	}
	os.WriteFile(cm.file, content, 0644)
}

type pollResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Code         int    `json:"code"`
		Message      string `json:"message"`
		RefreshToken string `json:"refresh_token"`
	} `json:"data"`
}

func HijackPollAPI(url string, resp *Response) {
	const api = `https://passport.bilibili.com/x/passport-login/web/qrcode/poll`
	if !strings.HasPrefix(url, api) {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return
	}
	var result pollResult
	err := json.Unmarshal(resp.Body, &result)
	if err != nil {
		return
	}
	if result.Data.Code != 0 {
		return
	}

	var cookies []*http.Cookie
	for _, kvs := range resp.Headers {
		if len(kvs) != 2 {
			continue
		}
		if !strings.EqualFold(kvs[0], "Set-Cookie") {
			continue
		}

		if cookie, err := http.ParseSetCookie(kvs[1]); err == nil {
			cookies = append(cookies, cookie)
		}
	}

	globalStorageManager.cookies.SetCookies(cookies, result.Data.RefreshToken)
}

func HijackLogout(rawurl string, req *http.Request) *http.Request {
	const api = "https://passport.bilibili.com/login/exit/v2"
	if !strings.HasPrefix(rawurl, api) {
		return req
	}
	csrf := GetCSRF()
	if csrf == "" {
		return req
	}

	form := make(url.Values)
	form.Add("gourl", "https://www.bilibili.com/")
	form.Add("biliCSRF", csrf)
	r, e := http.NewRequest(http.MethodPost, rawurl, strings.NewReader(form.Encode()))
	if e != nil {
		return req
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

func decodeRespCode(resp *Response) int {
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode
	}
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	err := json.Unmarshal(resp.Body, &result)
	if err != nil {
		return http.StatusInternalServerError
	}
	return result.Code
}

func HijackLogoutResp(rawurl string, resp *Response) {
	const api = "https://passport.bilibili.com/login/exit/v2"
	if !strings.HasPrefix(rawurl, api) {
		return
	}
	code := decodeRespCode(resp)
	if code == 0 || code == notLoginCode {
		globalStorageManager.cookies.clear()
	}
}

const notLoginCode = -101

func HijackIsCookieExpire(resp *Response) {
	if resp.StatusCode != http.StatusOK {
		return
	}
	code := decodeRespCode(resp)
	if code == notLoginCode {
		globalStorageManager.cookies.clear()
	}
}

func GetCSRF() string {
	return globalStorageManager.cookies.GetCSRF()
}

func (cm *cookiesManager) GetCSRF() string {
	for _, cookie := range cm.cookies {
		if cookie.Name == "bili_jct" {
			return cookie.Value
		}
	}
	return ""
}

func (cm *cookiesManager) ServeLogined(w http.ResponseWriter, r *http.Request) {
	cm.clearExpiredCookies()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(len(cm.cookies) > 0)
}

func (cm *cookiesManager) onStatusChanged() {
	for _, ch := range cm.wait {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (cm *cookiesManager) allocChan() (<-chan struct{}, func()) {
	ch := make(chan struct{})
	cm.mut.Lock()
	defer cm.mut.Unlock()

	id := cm.id
	cm.id++
	cm.wait[id] = ch

	return ch, func() {
		cm.mut.Lock()
		defer cm.mut.Unlock()
		delete(cm.wait, id)
	}
}

func (cm *cookiesManager) ServePoll(w http.ResponseWriter, r *http.Request) {
	wait, cancel := cm.allocChan()
	defer cancel()

	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	select {
	case <-wait:
	case <-timer.C:
	}

	cm.ServeLogined(w, r)
}

type storageManager struct {
	dir string

	search    *searchManager
	playlist  *playlistManager
	playstate *playstateManager
	settings  *settingsManager
	cookies   *cookiesManager
}

var globalStorageManager *storageManager

// must after `InitAudioCacheManager` and `InitImageCacheManager`
func InitStorageManager(dataDir string) {
	globalStorageManager = &storageManager{
		dir: dataDir,
		search: &searchManager{
			file: filepath.Join(dataDir, "search.json"),
			lru:  lru.NewWithMutex[string, bool](10),
		},
		playlist: &playlistManager{
			file: filepath.Join(dataDir, "playlist.json"),
		},
		playstate: &playstateManager{
			file: filepath.Join(dataDir, "playstate.json"),
		},
		settings: &settingsManager{
			file: filepath.Join(dataDir, "settings.json"),
		},
		cookies: &cookiesManager{
			file: filepath.Join(dataDir, "cookies.json"),
		},
	}

	// load
	globalStorageManager.search.load()
	globalStorageManager.playlist.load()
	globalStorageManager.playstate.load()
	globalStorageManager.settings.load()
	globalStorageManager.cookies.load()

	SetAudioCacheSizeLimit(globalStorageManager.settings.Audio * 1024 * 1024)
	SetImageCacheSizeLimit(globalStorageManager.settings.Image * 1024 * 1024)
}

func RouteStoreAPI(w http.ResponseWriter, r *http.Request) {
	// "search", "playlist", ("playmode", "playindex"), "settings", "login"

	path := strings.TrimPrefix(r.URL.Path, "/store/")
	switch path {
	case "search":
		globalStorageManager.search.ServeHTTP(w, r)

	case "play/list":
		globalStorageManager.playlist.ServeHTTP(w, r)

	case "play/state":
		globalStorageManager.playstate.ServeHTTP(w, r)

	case "settings":
		globalStorageManager.settings.ServeHTTP(w, r)

	case "cookie/logined":
		globalStorageManager.cookies.ServeLogined(w, r)
	case "cookie/poll":
		globalStorageManager.cookies.ServePoll(w, r)

	default:
		http.NotFound(w, r)
	}
}
