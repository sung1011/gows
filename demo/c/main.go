// ping
// echo
// push
// exec
// bin
package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sung1011/gows/demo/util"
)

var addr = flag.String("addr", "127.0.0.1:8081", "http service address")

const (
	connWait = 55 * time.Second

	pingLoopInterval = 1 * time.Second
	echoLoopInterval = 2 * time.Second
	pushLoopInterval = 3 * time.Second
)

func Run(method string, c *websocket.Conn) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	c.SetPongHandler(func(appData string) error {
		startTime, err := strconv.Atoi(appData)
		if err != nil {
			log.Fatal("err response data time()", err)
		}
		fmt.Printf("output: ping time=%v ms\n", util.GetMillisecTime()-startTime)
		return nil
	})

	go func() {
		if method == "exec" {
			cmd := strings.Join(os.Args[3:], " ")
			if err := util.WsWrite(c, util.MsgTypeExec, []byte(cmd)); err != nil {
				log.Fatal("err exec write", err)
			}
		}
		pingTicker := time.NewTicker(pingLoopInterval)
		echoTicker := time.NewTicker(echoLoopInterval)
		pushTicker := time.NewTicker(pushLoopInterval)
		for {
			select {
			case <-pingTicker.C:
				if method != "ping" {
					break
				}
				if err := util.WsWrite(c, websocket.PingMessage, []byte(util.GetMillisecTimeStr())); err != nil {
					log.Fatal("err ping write", err)
				}
			case <-echoTicker.C:
				if method != "echo" {
					break
				}
				var msg []byte
				if len(os.Args) <= 2 {
					msg = []byte("echo")
				} else {
					msg = []byte(strings.Join(os.Args[2:], " "))
				}
				if err := util.WsWrite(c, websocket.TextMessage, msg); err != nil {
					log.Fatal("err echo write", err)
				}
			case <-pushTicker.C:
				if method != "push" {
					break
				}
				if len(os.Args) <= 3 {
					log.Fatal("err param need [uid] push [pushMsg] [uid...]")
				}
				uids := strings.Join(os.Args[3:], " ")
				if err := util.WsWrite(c, util.MsgTypeJson, []byte(uids)); err != nil {
					log.Fatal("err push write", err)
				}
			case <-sig:
				if err := util.WsWrite(c, websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
					log.Fatal("err send close:", err)
				}
			}
		}
	}()
	m := make(chan []byte)
	go func() {
		for {
			msgType, msg, err := util.WsRead(c)
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				fmt.Println("websocket close")
				os.Exit(1)
			}
			if err != nil {
				fmt.Println("read err: ", err)
				os.Exit(1)
			}
			m <- msg
			switch msgType {
			case util.MsgTypeExec:
				os.Exit(0)
			}
		}
	}()

	for {
		fmt.Println("output:", string(<-m))
	}
}

func main() {
	flag.Parse()
	uid := "0"
	method := "echo"
	if len(os.Args) < 3 {
		log.Println("notice: need params [uid] [method], default", uid, method)
	} else {
		uid = os.Args[1]
		method = os.Args[2]
	}

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/" + method, RawQuery: "uid=" + uid}
	fmt.Println("url:", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}
	defer c.Close()
	c.SetReadDeadline(time.Now().Add(connWait))

	Run(method, c)
}
