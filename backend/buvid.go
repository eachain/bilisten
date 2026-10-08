package backend

import (
	"encoding/json"
	"net/http"
	"time"
)


type BuvId struct {
	b3 string
	b4 string
	t  time.Time
}

var buvIdManager = new(BuvId)

func (b *BuvId) AddCookies(r *http.Request) {
	if r.URL.String() == fingerURL {
		return
	}
	if b.b3 == "" || b.b4 == "" || time.Since(b.t) > time.Hour {
		b.update()
	}
	if b.b3 != "" {
		r.AddCookie(&http.Cookie{
			Name:  "buvid3",
			Value: b.b3,
		})
	}
	if b.b4 != "" {
		r.AddCookie(&http.Cookie{
			Name:  "buvid4",
			Value: b.b4,
		})
	}
}

const fingerURL = "https://api.bilibili.com/x/frontend/finger/spi"

func (b *BuvId) update() {
	now := time.Now()
	resp := DoRequestAPI(&Request{
		Tx:  NewTx(),
		URL: fingerURL,
	})
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			BuvId3 string `json:"b_3"`
			BuvId4 string `json:"b_4"`
		} `json:"data"`
	}
	err := json.Unmarshal(resp.Body, &result)
	if err != nil {
		return
	}
	if result.Code != 0 {
		return
	}
	b.b3 = result.Data.BuvId3
	b.b4 = result.Data.BuvId4
	b.t = now
}
