package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
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
	fmt.Println("CONN: success", r.URL.String())
	if err != nil {
		log.Fatal("Upgrade: ", err)
	}
	defer c.Close()

	c.SetReadDeadline(time.Now().Add(connWait))
	c.SetPingHandler(func(appData string) error {
		err := util.WsWrite(c, websocket.PongMessage, ([]byte(appData)))
		if err != nil {
			fmt.Println("ERR: set pong", err)
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
			fmt.Println("CONN: disconn, uid", uid)
			return
		}
		if err != nil {
			fmt.Println("ERR: server read", msgType, msg, err)
			return
		}
		switch msgType {
		case util.MsgTypeJson:
			j := &util.JsonRPC{}
			err = json.Unmarshal(msg, j)
			// TODO 根据 json 的 method 路由到不同的handler 现在只有push
			if j.Method == "push" {
				fmt.Println("PUSH:", string(msg))
				for _, pv := range j.Params {
					kv := strings.Split(pv, "=")
					if kv[0] != "uids" {
						continue
					}
					if kv[1] == "" { // push all
						toCAll := bd.GetConnAll()
						for _, toC := range toCAll {
							err = util.WsWrite(toC, websocket.TextMessage, []byte("push by "+uid))
						}
					} else { // push one
						toC, cExists := bd.GetConn(kv[1])
						if cExists == false {
							continue
						}
						err = util.WsWrite(toC, websocket.TextMessage, []byte("push by "+uid))
					}
				}
			}
		case util.MsgTypeExec:
			fmt.Println("EXEC:", string(msg))
			tmp := strings.Split(string(msg), " ")
			bin := tmp[0]
			bin, err = exec.LookPath(bin)
			if err != nil {
				break
			}
			var args []string
			if len(tmp) > 1 {
				args = tmp[1:]
			} else {
				args = []string{}
			}
			cmd := exec.Command(bin, args...)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				break
			}
			defer stdout.Close()

			err = cmd.Start()
			if err != nil {
				break
			}

			s := bufio.NewScanner(stdout)
			for s.Scan() {
				if err := util.WsWrite(c, websocket.TextMessage, s.Bytes()); err != nil {
					c.Close()
					break
				}
			}
			err = cmd.Wait()
			if err != nil {
				break
			}
			if s.Err() != nil {
				log.Println("scan:", s.Err())
			}
			defer c.Close()
			return

		default:
			fmt.Println("ECHO:", string(msg))
			err = util.WsWrite(c, msgType, msg)
		}
		if err != nil {
			fmt.Println("ERR: write", err)
			return
		}
	}
}

func main() {
	flag.Parse()
	// TODO Run拆开
	http.HandleFunc("/ping", Run)
	http.HandleFunc("/echo", Run)
	http.HandleFunc("/push", Run)
	http.HandleFunc("/exec", Run)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
