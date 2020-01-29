package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sung1011/gows/demo/util"
)

var addr = flag.String("addr", "127.0.0.1:8081", "http service address")

const (
	connWait = 60 * time.Second
)

func Run(w http.ResponseWriter, r *http.Request) {
	urg := websocket.Upgrader{}
	c, err := urg.Upgrade(w, r, nil)
	fmt.Println("conn success", r.URL.String())
	if err != nil {
		log.Fatal("upgrade: ", err)
	}
	defer c.Close()

	c.SetReadDeadline(time.Now().Add(connWait))
	c.SetPingHandler(func(appData string) error {
		err := util.WsWrite(c, websocket.PongMessage, ([]byte(appData)))
		if err != nil {
			fmt.Println("err: set pong", err)
		}
		return err
	})

	bd := util.GetBinderInstance()
	uid := r.URL.Query()["uid"][0]
	bd.Bind(uid, c)
	defer bd.Unbind(uid)
	for {
		msgType, msg, err := util.WsRead(c)
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			bd.Unbind(uid)
			fmt.Println("disconn, uid", uid)
			return
		}
		if err != nil {
			fmt.Println("err: server read: ", msgType, msg, err)
			return
		}
		switch msgType {
		case util.MsgTypeJson:
			j := &util.JsonRPC{}
			msgData := util.GetMsgData(msg)
			err = json.Unmarshal(msgData, j)
			// TODO 根据 json 的 method 路由到不同的handler
			if j.Method == "push" {
				for _, toUID := range j.Params {
					toC, cExists := bd.GetConn(toUID)
					if cExists == false {
						continue
					}
					err = util.WsWrite(toC, websocket.TextMessage, []byte("push by "+uid))
				}
			}

		default:
			err = util.WsWrite(c, msgType, msg)
		}
		if err != nil {
			fmt.Println("err: write", err)
			return
		}
	}
}

func main() {
	flag.Parse()
	// TODO Run拆开
	http.HandleFunc("/run", Run)
	http.HandleFunc("/ping", Run)
	http.HandleFunc("/echo", Run)
	http.HandleFunc("/push", Run)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
