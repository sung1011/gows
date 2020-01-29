package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (
	MsgTypeJson = 21
	MsgTypeCypt = 22
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
		// TODO
		j := &JsonRPC{Id: "123", Method: "push", Params: []string{"4", "5", "6"}}
		err = writeJSON(c, strconv.Itoa(MsgTypeJson), j)
	default:
		bb := bytes.Buffer{}
		bb.Write([]byte(msgTypeStr))
		bb.Write(msg)
		msg = bb.Bytes()
		err = c.WriteMessage(msgType, msg)
	}
	// fmt.Println(msgType, string(msg))
	return err
}

func WsRead(c *websocket.Conn) (int, []byte, error) {
	var err error
	msgType, msg, err := c.ReadMessage()
	// fmt.Println("---", string(msg))
	if len(msg) == 0 || isControl(msgType) {
		return msgType, msg, err
	}
	switch string(msg[0:2]) {
	case strconv.Itoa(MsgTypeCypt):
	case strconv.Itoa(MsgTypeJson):
		msgType = MsgTypeJson
	default:
		msg = GetMsgData(msg)
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
