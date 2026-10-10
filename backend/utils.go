package backend

import (
	"crypto/rand"
	"encoding/hex"
)

type TimeoutError interface {
	Timeout() bool
}

func NewTx() string {
	var p [4]byte
	rand.Read(p[:])
	return hex.EncodeToString(p[:])
}
