package backend

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eachain/lru"
)

const AudiosDir = "audios"

type SegmentBase struct {
	Initialization string `json:"Initialization"`
	IndexRange     string `json:"indexRange"`
}

type AudioCacheInfo struct {
	Bvid        string      `json:"bvid"`
	Cid         string      `json:"cid"`
	Title       string      `json:"title"`
	Pic         string      `json:"pic"`
	Author      string      `json:"author"`
	Mid         string      `json:"mid"`
	Id          int         `json:"id"` // SoundQuality
	Duration    int64       `json:"duration"`
	Size        int64       `json:"size"`
	Bandwidth   int         `json:"bandwidth"`
	MimeType    string      `json:"mime_type"` // "audio/mp4"
	Codecs      string      `json:"codecs"`    // "mp4a.40.2"
	SegmentBase SegmentBase `json:"SegmentBase"`
	Qualities   []string    `json:"qs,omitempty"`
}

type audioCacheManager struct {
	dir   string
	mut   sync.RWMutex
	aof   *os.File
	lines int
	lru   *lru.LRU[string, int64]
	size  int64
	limit int64
}

var globalAudioCacheManager *audioCacheManager

func InitAudioCacheManager(dataDir, cacheDir string) {
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		panic(err) // TODO
	}
	aof, err := os.OpenFile(filepath.Join(dataDir, "audios.aof"), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err) // TODO
	}
	globalAudioCacheManager = &audioCacheManager{
		dir: filepath.Join(cacheDir, AudiosDir),
		aof: aof,
		lru: lru.New[string, int64](0),
	}

	sc := bufio.NewScanner(aof)
	for sc.Scan() {
		line := sc.Bytes()
		globalAudioCacheManager.lines++
		op, name, size, ok := parseCacheLine(line)
		if !ok {
			continue
		}
		if op == "d" {
			globalAudioCacheManager.lru.Remove(name)
		} else {
			globalAudioCacheManager.lru.Set(name, size)
		}
	}
	if err := sc.Err(); err != nil {
		log.Printf("audio cache manager: scan: %v", err)
		_, err = aof.Seek(0, io.SeekEnd)
		if err != nil {
			log.Printf("audio cache manager: seek to eof: %v", err)
		}
	}

	var total int64
	for _, size := range globalAudioCacheManager.lru.All() {
		total += size
	}
	globalAudioCacheManager.size = total
}

func (acm *audioCacheManager) GetUsedSize() int64 {
	acm.mut.RLock()
	defer acm.mut.RUnlock()
	return acm.size
}

func (acm *audioCacheManager) writeAOF(op string, name string, size int64) {
	fmt.Fprintf(acm.aof, "%v,%v,%v\n", op, name, size)
	acm.lines++

	if acm.lru.Len() > 0 && acm.lines >= 1000 && acm.lines >= acm.lru.Len()*2 { // 即有效内容仅占50%
		var buf bytes.Buffer
		acm.lines = 0
		for file, size := range acm.lru.All() {
			fmt.Fprintf(&buf, "%v,%v,%v\n", "r", file, size)
			acm.lines++
		}
		acm.aof.Truncate(0)
		acm.aof.Seek(0, io.SeekStart)
		acm.aof.Write(buf.Bytes())
	}
}

func (acm *audioCacheManager) evict(tx string) {
	if acm.limit <= 0 {
		return
	}
	var files []string
	for acm.size > acm.limit {
		file, size, ok := acm.lru.RemoveOldest()
		if !ok {
			break
		}
		files = append(files, file)
		acm.size -= size
	}
	if len(files) == 0 {
		return
	}

	go func() {
		for _, file := range files {
			acm.mut.Lock()
			acm.writeAOF("d", file, 0)
			acm.mut.Unlock()
			acm.removeFile(tx, file)
		}
	}()
}

func (acm *audioCacheManager) remove(tx string, name string) {
	acm.mut.Lock()
	size, ok := acm.lru.Remove(name)
	if ok {
		acm.size -= size
		acm.writeAOF("d", name, size)
	}
	acm.mut.Unlock()
	go acm.removeFile(tx, name)
}

func (acm *audioCacheManager) removeFile(tx string, file string) {
	audio := filepath.Join(acm.dir, file+".m4s")
	meta := filepath.Join(acm.dir, file+".json")
	tmp := filepath.Join(acm.dir, file)
	for i := range 3 {
		err := os.RemoveAll(audio)
		if err != nil {
			log.Printf("%v audio cache: remove %q: %v", tx, audio, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		err = os.RemoveAll(tmp)
		if err != nil {
			log.Printf("%v audio cache: remove %q: %v", tx, tmp, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		err = os.RemoveAll(meta)
		if err != nil {
			log.Printf("%v audio cache: remove %q: %v", tx, meta, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
	}
}

func SetAudioCacheSizeLimit(limit int64) {
	globalAudioCacheManager.SetSizeLimit(limit)
}

func (acm *audioCacheManager) SetSizeLimit(limit int64) {
	tx := NewTx()

	acm.mut.Lock()
	defer acm.mut.Unlock()
	acm.limit = limit
	acm.evict(tx)
}

func ClearAudioCache() {
	globalAudioCacheManager.Clear()
}

// 清空缓存
func (acm *audioCacheManager) Clear() {
	acm.mut.Lock()
	defer acm.mut.Unlock()

	os.RemoveAll(acm.dir)

	acm.aof.Truncate(0)
	acm.aof.Seek(0, io.SeekStart)
	acm.lines = 0
	acm.size = 0
	acm.lru = lru.New[string, int64](0)
}

func (acm *audioCacheManager) Write(tx string, name, file string, data []byte, total int64) {
	err := os.WriteFile(file, data, 0644)
	if err != nil {
		log.Printf("%v cache audio: write %q: %v", tx, file, err)
		go acm.remove(tx, name)
		return
	}

	log.Printf("%v cache audio: %q", tx, file)

	partSize := int64(len(data))
	acm.mut.Lock()
	defer acm.mut.Unlock()
	oriSize, ok := acm.lru.Get(name)
	if !ok {
		go acm.remove(tx, name)
		return
	}
	size := oriSize + partSize
	acm.lru.Set(name, size)
	acm.size += partSize
	acm.writeAOF("w", name, size)
	if size == total {
		go acm.merge(tx, name)
	} else if size >= total {
		go acm.remove(tx, name)
	}
	acm.evict(tx)
}

type audioPart struct {
	name  string
	start int64
	end   int64
}

func (acm *audioCacheManager) merge(tx string, name string) {
	var success bool
	defer func() {
		if !success {
			acm.remove(tx, name)
		}
	}()

	tmpDir := filepath.Join(acm.dir, name)
	meta := filepath.Join(tmpDir, name+".json")
	content, err := os.ReadFile(meta)
	if err != nil {
		log.Printf("%v merge %v, read meta %q: %v", tx, name, meta, err)
		return
	}
	info := new(AudioCacheInfo)
	err = json.Unmarshal(content, info)
	if err != nil {
		log.Printf("%v merge %v, json loads meta %q: %v", tx, name, meta, err)
		return
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		log.Printf("%v merge %v, read tmp dir %q: %v", tx, name, tmpDir, err)
		return
	}

	prefix := name + ".m4s."
	parts := make([]audioPart, 0, len(entries)-1)
	for _, entry := range entries {
		base := filepath.Base(entry.Name())
		if !strings.HasPrefix(base, prefix) {
			continue
		}
		partRange := strings.TrimPrefix(base, prefix)
		startStr, endStr, ok := strings.Cut(partRange, "-")
		if !ok {
			continue
		}
		start, err := strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			log.Printf("%v merge %v, parse range start of %q: %v", tx, name, base, err)
			return
		}
		end, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			log.Printf("%v merge %v, parse range end of %q: %v", tx, name, base, err)
			return
		}
		parts = append(parts, audioPart{
			name:  base,
			start: start,
			end:   end,
		})
	}

	slices.SortFunc(parts, func(a, b audioPart) int {
		if a.start < b.start {
			return -1
		}
		return 1
	})

	var end int64 = -1
	for _, part := range parts {
		if part.start != end+1 {
			log.Printf("%v merge %v, range not found: %v-%v", tx, name, end, part.start)
			return
		}
		end = part.end
	}
	if end != info.Size-1 {
		log.Printf("%v merge %v, range not found: %v-%v", tx, name, end, info.Size-1)
		return
	}

	mergeFile := filepath.Join(tmpDir, name+".m4s")
	mergeFp, err := os.OpenFile(mergeFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("%v merge %v, open merge file %q: %v", tx, name, mergeFile, err)
		return
	}
	var mergeFpClosed bool
	defer func() {
		if !mergeFpClosed {
			mergeFp.Close()
		}
	}()

	for _, part := range parts {
		var failErr error
		func() {
			partFile := filepath.Join(tmpDir, part.name)
			partFp, err := os.Open(partFile)
			if err != nil {
				log.Printf("%v merge %v, open part file %q: %v", tx, name, partFile, err)
				failErr = err
				return
			}
			defer partFp.Close()

			_, err = partFp.WriteTo(mergeFp)
			if err != nil {
				log.Printf("%v merge %v, part file %q write to merge file: %v", tx, name, partFile, err)
				failErr = err
				return
			}
		}()
		if failErr != nil {
			return
		}
	}

	mergeFp.Close()
	mergeFpClosed = true

	dstMeta := filepath.Join(acm.dir, name+".json")
	dstAudio := filepath.Join(acm.dir, name+".m4s")
	err = os.Rename(meta, dstMeta)
	if err != nil {
		log.Printf("%v merge %v, rename meta file %q->%q: %v", tx, name, meta, dstMeta, err)
		return
	}
	err = os.Rename(mergeFile, dstAudio)
	if err != nil {
		log.Printf("%v merge %v, rename audio file %q->%q: %v", tx, name, mergeFile, dstAudio, err)
		return
	}

	log.Printf("%v merge %v ok", tx, name)

	success = true

	for range 5 {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			log.Printf("%v merge %v, remove tmp dir %q: %v", tx, name, tmpDir, err)
			time.Sleep(time.Second)
			continue
		}
	}
}

func (acm *audioCacheManager) genUniqName(bvid, cid, q string) string {
	return fmt.Sprintf("%v-%v-%v", bvid, cid, q)
}

func (acm *audioCacheManager) GetAudioInfo(bvid, cid, q string) *AudioCacheInfo {
	tx := NewTx()

	name := acm.genUniqName(bvid, cid, q)

	var ok bool
	acm.mut.Lock()
	var first string
	for first = range acm.lru.All() {
		break
	}
	ok = name == first
	if !ok {
		var size int64
		size, ok = acm.lru.Get(name)
		if ok {
			acm.writeAOF("r", name, size)
		}
	}
	acm.mut.Unlock()

	if !ok {
		return nil
	}

	path := filepath.Join(acm.dir, name+".json")
	content, err := os.ReadFile(path)
	if err == nil {
		info := new(AudioCacheInfo)
		err = json.Unmarshal(content, info)
		if err != nil {
			log.Printf("%v json load meta %q: %v", tx, path, err)
			acm.remove(tx, name)
			return nil
		}
		return info
	}
	if !os.IsNotExist(err) {
		log.Printf("%v read meta %q: %v", tx, path, err)
		acm.remove(tx, name)
		return nil
	}

	path = filepath.Join(acm.dir, name, name+".json")
	content, err = os.ReadFile(path)
	if err != nil {
		acm.remove(tx, name)
		return nil
	}

	info := new(AudioCacheInfo)
	err = json.Unmarshal(content, info)
	if err != nil {
		log.Printf("%v json load meta %q: %v", tx, path, err)
		acm.remove(tx, name)
		return nil
	}
	return info
}

func (acm *audioCacheManager) getLimit() int64 {
	acm.mut.RLock()
	defer acm.mut.RUnlock()
	return acm.limit
}

func (acm *audioCacheManager) SetAudioInfo(tx string, info *AudioCacheInfo) {
	name := acm.genUniqName(info.Bvid, info.Cid, SoundQuality[info.Id])

	content, err := json.Marshal(info)
	if err != nil {
		log.Printf("%v json dumps audio info: %v", tx, err)
		return
	}
	size := int64(len(content))
	if size > acm.getLimit() {
		return
	}

	dir := filepath.Join(acm.dir, name)
	os.MkdirAll(dir, 0755)
	file := filepath.Join(dir, name+".json")
	err = os.WriteFile(file, content, 0644)
	if err != nil {
		go acm.remove(tx, name)
		log.Printf("%v write audio info to %q: %v", tx, file, err)
		return
	}

	acm.mut.Lock()
	_, ok := acm.lru.Get(name)
	if !ok {
		acm.lru.Set(name, size)
		acm.size += size
		acm.writeAOF("w", name, size)
	}
	acm.mut.Unlock()
}

func (acm *audioCacheManager) correctAudioSizeByRange(tx string, info *AudioCacheInfo, contentRange string) {
	if info.Size > 0 {
		return
	}

	// Content-Range: bytes 6481-83319/32419841
	_, after, ok := strings.Cut(contentRange, "/")
	if !ok {
		return
	}
	size, err := strconv.ParseInt(after, 10, 64)
	if err != nil {
		return
	}

	acm.mut.RLock()
	if size > acm.limit {
		acm.mut.RUnlock()
		name := acm.genUniqName(info.Bvid, info.Cid, SoundQuality[info.Id])
		acm.remove(tx, name)
		return
	}
	acm.mut.RUnlock()

	info.Size = size
	content, err := json.Marshal(info)
	if err != nil {
		return
	}

	name := acm.genUniqName(info.Bvid, info.Cid, SoundQuality[info.Id])
	meta := filepath.Join(acm.dir, name, name+".json")
	err = os.WriteFile(meta, content, 0644)
	if err != nil {
		return
	}

	acm.mut.Lock()
	defer acm.mut.Unlock()
	value, ok := acm.lru.Get(name)
	if ok && value == int64(len(content)+1-len(after)) {
		acm.lru.Set(name, int64(len(content)))
		acm.size += int64(len(after) - 1)
		acm.writeAOF("w", name, int64(len(content)))
	}
}

func ProxyAudio(w http.ResponseWriter, r *http.Request) {
	globalAudioCacheManager.ProxyAudio(w, r)
}

func AudioCacheEtag(bvid, cid, q string) string {
	return fmt.Sprintf("audio-%v-%v-%v", bvid, cid, q)
}

func parseBvidCidQuality(path string) (bvid, cid, q string, ok bool) {
	fields := strings.Split(path, "/")
	if len(fields) < 4 {
		return
	}
	// "/audio/bvid/cid/high.m4s"
	bvid = fields[2]
	cid = fields[3]
	if len(fields) > 4 {
		q, _, _ = strings.Cut(fields[4], ".")
	}
	ok = true
	return
}

func (acm *audioCacheManager) ProxyAudio(w http.ResponseWriter, r *http.Request) {
	if ReuseCache(w, r) {
		return
	}
	tx := NewTx()

	bvid, cid, q, ok := parseBvidCidQuality(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	viaCDN := func() {
		info, status := getAudioInfo(tx, bvid, cid, []string{q})
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		ProxyCDN(w, r, CDNOptions{
			Tx:          tx,
			URL:         info.BaseURL,
			ContentType: ContentTypeOf(info.MimeType, info.Codecs),
			Etag:        AudioCacheEtag(bvid, cid, q),
		})
	}

	name := acm.genUniqName(bvid, cid, q)

	acm.mut.Lock()
	_, ok = acm.lru.Get(name)
	acm.mut.Unlock()
	if !ok {
		viaCDN()
		return
	}

	file := filepath.Join(acm.dir, name+".m4s")
	_, err := os.Stat(file)
	if err == nil {
		http.ServeFile(w, r, file)
		return
	}

	meta := filepath.Join(acm.dir, name, name+".json")
	content, err := os.ReadFile(meta)
	if err != nil {
		log.Printf("%v read meta file %q: %v", tx, meta, err)
		go acm.remove(tx, name)
		viaCDN()
		return
	}

	info := new(AudioCacheInfo)
	err = json.Unmarshal(content, info)
	if err != nil {
		log.Printf("%v json loads meta from file %q: %v", tx, meta, err)
		go acm.remove(tx, name)
		viaCDN()
		return
	}

	reqRange := r.Header.Get("Range")
	suffix := strings.TrimPrefix(reqRange, "bytes=") // bytes=56349-124246
	tmp := filepath.Join(acm.dir, name, name+".m4s."+suffix)
	_, err = os.Stat(tmp)
	if err == nil {
		content, err := os.ReadFile(tmp)
		if err != nil {
			log.Printf("%v read part file %q: %v", tx, tmp, err)
			go acm.remove(tx, name)
			viaCDN()
			return
		}

		// Content-Range: bytes 56349-124246/3014805
		w.Header().Set("Content-Range", reqRange+"/"+strconv.FormatInt(info.Size, 10))
		w.Header().Set("Content-Type", ContentTypeOf(info.MimeType, info.Codecs))
		w.WriteHeader(http.StatusPartialContent)
		w.Write(content)
		return
	}

	audio, status := getAudioInfo(tx, bvid, cid, []string{q})
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}

	hook := acm.NewHook(tx, name, tmp, info)
	ProxyCDN(w, r, CDNOptions{
		Tx:             tx,
		URL:            audio.BaseURL,
		ContentType:    ContentTypeOf(info.MimeType, info.Codecs),
		Etag:           AudioCacheEtag(bvid, cid, q),
		HookRespHeader: hook.HookRespHeader,
		HookRespStatus: hook.HookRespStatus,
		HookRespBody:   hook.HookRespBody,
		HookRespOver:   hook.HookRespOver,
	})
}

func (acm *audioCacheManager) genName(url string) string {
	ext := ".webp"
	idx := strings.LastIndexByte(url, '.')
	if idx > 0 {
		ext = url[idx:]
	}
	sum := md5.Sum([]byte(url))
	name := hex.EncodeToString(sum[:])
	return name + ext
}

func (acm *audioCacheManager) NewHook(tx string, name, file string, info *AudioCacheInfo) *audioCacheHook {
	return &audioCacheHook{
		acm:  acm,
		tx:   tx,
		name: name,
		file: file,
		info: info,
	}
}

type audioCacheHook struct {
	acm  *audioCacheManager
	tx   string
	name string
	file string
	info *AudioCacheInfo

	size   int64
	status int
	body   []byte
}

func (hook *audioCacheHook) HookRespHeader(header http.Header) {
	hook.size, _ = strconv.ParseInt(header.Get("Content-Length"), 10, 64)

	hook.acm.correctAudioSizeByRange(hook.tx, hook.info, header.Get("Content-Range"))
}

func (hook *audioCacheHook) HookRespStatus(status int) {
	hook.status = status
}

func (hook *audioCacheHook) HookRespBody(body []byte) {
	if hook.status != http.StatusPartialContent {
		return
	}
	if hook.size <= 0 {
		return
	}

	hook.body = append(hook.body, body...)
}

func (hook *audioCacheHook) HookRespOver(err error) {
	if hook.status != http.StatusPartialContent {
		return
	}
	if hook.size <= 0 {
		return
	}
	if len(hook.body) != int(hook.size) {
		return
	}

	content, _ := json.Marshal(hook.info)
	go hook.acm.Write(hook.tx, hook.name, hook.file, hook.body, hook.info.Size+int64(len(content)))
}
