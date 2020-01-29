package util

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Binder struct {
	mu       sync.RWMutex
	uid2conn map[string]*websocket.Conn
}

var once sync.Once
var bd *Binder

func GetBinderInstance() *Binder {
	once.Do(func() {
		bd = &Binder{
			uid2conn: make(map[string]*websocket.Conn),
		}
	})
	return bd
}

func (bd *Binder) GetConn(uid string) (*websocket.Conn, bool) {
	c, ok := bd.uid2conn[uid]
	return c, ok
}

func (bd *Binder) Bind(uid string, conn *websocket.Conn) {
	bd.mu.Lock()
	defer bd.mu.Unlock()
	bd.uid2conn[uid] = conn
}

func (bd *Binder) Unbind(uid string) {
	bd.mu.Lock()
	defer bd.mu.Unlock()
	delete(bd.uid2conn, uid)
}
