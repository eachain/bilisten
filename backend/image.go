package backend

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eachain/lru"
)

const ImagesDir = "images"

type imageCacheManager struct {
	dir   string
	mut   sync.RWMutex
	aof   *os.File
	lines int
	lru   *lru.LRU[string, int64]
	size  int64
	limit int64
}

var globalImageCacheManager *imageCacheManager

func InitImageCacheManager(dataDir, cacheDir string) {
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		panic(err) // TODO
	}
	aof, err := os.OpenFile(filepath.Join(dataDir, "images.aof"), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err) // TODO
	}
	globalImageCacheManager = &imageCacheManager{
		dir: filepath.Join(cacheDir, ImagesDir),
		aof: aof,
		lru: lru.New[string, int64](0),
	}

	sc := bufio.NewScanner(aof)
	for sc.Scan() {
		line := sc.Bytes()
		globalImageCacheManager.lines++
		op, name, size, ok := parseCacheLine(line)
		if !ok {
			continue
		}
		if op == "d" {
			size, ok := globalImageCacheManager.lru.Remove(name)
			if ok {
				globalImageCacheManager.size -= size
			}
		} else {
			if old, ok := globalImageCacheManager.lru.Get(name); ok {
				globalImageCacheManager.size -= old
			}
			globalImageCacheManager.lru.Set(name, size)
			globalImageCacheManager.size += size
		}
	}
	if err := sc.Err(); err != nil {
		log := NewLogger("InitImageCacheManager")
		log.Error("scan: %v", err)
		_, err = aof.Seek(0, io.SeekEnd)
		if err != nil {
			log.Error("seek to eof: %v", err)
		}
	}
}

func parseCacheLine(line []byte) (op string, name string, size int64, ok bool) {
	fields := bytes.Split(line, []byte{','})
	if len(fields) < 3 {
		return
	}
	size, err := strconv.ParseInt(string(fields[2]), 10, 64)
	if err != nil {
		return
	}
	op = string(fields[0])
	name = string(fields[1])
	ok = true
	return
}

func (icm *imageCacheManager) writeAOF(op string, name string, size int64) {
	fmt.Fprintf(icm.aof, "%v,%v,%v\n", op, name, size)
	icm.lines++

	if icm.lru.Len() > 0 && icm.lines >= 100 && icm.lines >= icm.lru.Len()*2 { // 即有效内容仅占50%
		var buf bytes.Buffer
		icm.lines = 0
		for file, size := range icm.lru.Backward() {
			fmt.Fprintf(&buf, "%v,%v,%v\n", "r", file, size)
			icm.lines++
		}
		icm.aof.Truncate(0)
		icm.aof.Seek(0, io.SeekStart)
		icm.aof.Write(buf.Bytes())
	}
}

func (icm *imageCacheManager) evict(tx string) {
	if icm.limit <= 0 {
		return
	}
	files := make(map[string]int64)
	for icm.size > icm.limit {
		file, size, ok := icm.lru.RemoveOldest()
		if !ok {
			break
		}
		files[file] = size
		icm.size -= size
	}
	if len(files) == 0 {
		return
	}

	go func() {
		for file, size := range files {
			icm.mut.Lock()
			icm.writeAOF("d", file, size)
			icm.mut.Unlock()

			log := NewLogger("images:evict").Tx(tx)
			path := filepath.Join(icm.dir, file)
			for i := range 3 {
				err := os.RemoveAll(path)
				if err != nil {
					log.Error("evict %q: %v", path, err)
					time.Sleep(time.Second * time.Duration(i+1))
					continue
				}
			}
		}
	}()
}

func SetImageCacheSizeLimit(limit int64) {
	globalImageCacheManager.SetSizeLimit(limit)
}

func (icm *imageCacheManager) SetSizeLimit(limit int64) {
	tx := NewTx()

	icm.mut.Lock()
	defer icm.mut.Unlock()
	icm.limit = limit
	icm.evict(tx)
}

func (icm *imageCacheManager) GetUsedSize() int64 {
	icm.mut.RLock()
	defer icm.mut.RUnlock()
	return icm.size
}

func ClearImageCache() {
	globalImageCacheManager.Clear()
}

// 清空缓存
func (icm *imageCacheManager) Clear() {
	icm.mut.Lock()
	defer icm.mut.Unlock()

	os.RemoveAll(icm.dir)

	icm.aof.Truncate(0)
	icm.aof.Seek(0, io.SeekStart)
	icm.lines = 0
	icm.size = 0
	icm.lru = lru.New[string, int64](0)
}

func (icm *imageCacheManager) getLimit() int64 {
	icm.mut.RLock()
	defer icm.mut.RUnlock()
	return icm.limit
}

func (icm *imageCacheManager) Write(tx string, file string, data []byte) {
	size := int64(len(data))
	if size > icm.getLimit() {
		return
	}

	log := NewLogger("WriteImage").Tx(tx)
	err := os.MkdirAll(icm.dir, 0755)
	if err != nil {
		log.Error("mkdir %q: %v", icm.dir, err)
		return
	}

	path := filepath.Join(icm.dir, file)
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		log.Error("write %q: %v", path, err)
		return
	}

	log.Info("cache image: %q done", path)

	icm.mut.Lock()
	defer icm.mut.Unlock()
	if old, ok := icm.lru.Get(file); ok {
		icm.size -= old
	}
	icm.lru.Set(file, size)
	icm.size += size
	icm.writeAOF("w", file, size)
	icm.evict(tx)
}

func ProxyImage(w http.ResponseWriter, r *http.Request) {
	globalImageCacheManager.ProxyImage(w, r)
}

func (icm *imageCacheManager) ProxyImage(w http.ResponseWriter, r *http.Request) {
	if ReuseCache(w, r) {
		return
	}

	tx := NewTx()
	url := r.URL.Query().Get("url")
	file := icm.genName(url)

	viaCDN := func() {
		hook := icm.NewHook(tx, file)
		ProxyCDN(w, r, CDNOptions{
			Tx:             tx,
			URL:            url,
			HookRespHeader: hook.HookRespHeader,
			HookRespStatus: hook.HookRespStatus,
			HookRespBody:   hook.HookRespBody,
			HookRespOver:   hook.HookRespOver,
		})
	}

	var ok bool
	icm.mut.Lock()
	var first string
	for first = range icm.lru.All() {
		break
	}
	ok = file == first
	if !ok {
		var size int64
		size, ok = icm.lru.Get(file)
		if ok {
			icm.writeAOF("r", file, size)
		}
	}
	icm.mut.Unlock()

	if !ok {
		viaCDN()
		return
	}

	path := filepath.Join(icm.dir, file)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			icm.mut.Lock()
			size, ok := icm.lru.Remove(file)
			if ok {
				icm.size -= size
				icm.writeAOF("d", file, size)
			}
			icm.mut.Unlock()
		}

		viaCDN()
		return
	}

	sum := md5.Sum(content)
	etag := base64.StdEncoding.EncodeToString(sum[:])
	w.Header().Set("etag", etag)
	http.ServeContent(w, r, file, time.Time{}, bytes.NewReader(content))
}

func (icm *imageCacheManager) genName(url string) string {
	ext := ".webp"
	idx := strings.LastIndexByte(url, '.')
	if idx > 0 {
		ext = url[idx:]
	}
	sum := md5.Sum([]byte(url))
	name := hex.EncodeToString(sum[:])
	return name + ext
}

func (icm *imageCacheManager) NewHook(tx string, file string) *imageCacheHook {
	return &imageCacheHook{
		icm:  icm,
		tx:   tx,
		file: file,
	}
}

type imageCacheHook struct {
	icm  *imageCacheManager
	tx   string
	file string

	size   int64
	status int
	body   []byte
}

func (hook *imageCacheHook) HookRespHeader(header http.Header) {
	hook.size, _ = strconv.ParseInt(header.Get("content-length"), 10, 64)
}

func (hook *imageCacheHook) HookRespStatus(status int) {
	hook.status = status
}

func (hook *imageCacheHook) HookRespBody(body []byte) {
	if hook.status != http.StatusOK {
		return
	}
	if hook.size <= 0 {
		return
	}

	hook.body = append(hook.body, body...)
}

func (hook *imageCacheHook) HookRespOver(err error) {
	if hook.status != http.StatusOK {
		return
	}
	if hook.size <= 0 {
		return
	}
	if len(hook.body) != int(hook.size) {
		return
	}

	go hook.icm.Write(hook.tx, hook.file, hook.body)
}
