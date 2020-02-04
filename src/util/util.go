package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	MsgTypeExec = 21 + iota
	MsgTypeJson
	MsgTypeCypt
)

func GetMillisecTime() int {
	return int(time.Now().UnixNano() / 1e6)
}

func GetMillisecTimeStr() string {
	return strconv.Itoa(GetMillisecTime())
}

func WsWrite(c *websocket.Conn, msgType int, msg []byte) error {
	if isControl(msgType) {
		return c.WriteMessage(msgType, msg)
	}
	msgTypeStr := fmt.Sprintf("%02d", msgType)
	var err error
	switch msgTypeStr {
	case strconv.Itoa(MsgTypeCypt):
	case strconv.Itoa(MsgTypeJson):
		var pushMsg, uids string
		if len(msg) >= 1 {
			s := string(msg)
			sm := strings.Split(s, " ")
			pushMsg = string(sm[0])
			uids = string(strings.Join(sm[1:], ","))
		} else {
			pushMsg = ""
			uids = ""
		}
		j := &JsonRPC{Id: "123", Method: "push", Params: []string{"pushMsg=" + pushMsg, "uids=" + uids}}
		err = writeJSON(c, strconv.Itoa(MsgTypeJson), j)
	default:
		bb := bytes.Buffer{}
		bb.Write([]byte(msgTypeStr))
		bb.Write(msg)
		msg = bb.Bytes()
		err = c.WriteMessage(websocket.TextMessage, msg)
	}
	return err
}

func WsRead(c *websocket.Conn) (int, []byte, error) {
	var err error
	msgType, msg, err := c.ReadMessage()
	// fmt.Println("---", string(msg))
	if len(msg) == 0 || isControl(msgType) {
		return msgType, msg, err
	}
	flag := string(msg[0:2])
	msg = GetMsgData(msg)
	switch flag {
	case strconv.Itoa(MsgTypeCypt):
	case strconv.Itoa(MsgTypeExec):
		msgType = MsgTypeExec
	case strconv.Itoa(MsgTypeJson):
		msgType = MsgTypeJson
	}
	// fmt.Println("------", string(msg))
	return msgType, bytes.Trim(msg, "\n"), err
}

type JsonRPC struct {
	Id     string   `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

func GetMsgData(b []byte) []byte {
	return b[2:]
}

func isControl(frameType int) bool {
	return frameType == websocket.CloseMessage || frameType == websocket.PingMessage || frameType == websocket.PongMessage
}

func writeJSON(c *websocket.Conn, flag string, v interface{}) error {
	w, err := c.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	w.Write([]byte(flag))
	err1 := json.NewEncoder(w).Encode(v)
	err2 := w.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func readJSON(c *websocket.Conn, v interface{}) error {
	_, r, err := c.NextReader()
	if err != nil {
		return err
	}
	err = json.NewDecoder(r).Decode(v)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return err
}
