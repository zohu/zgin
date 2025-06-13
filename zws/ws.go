package zws

import (
	"bytes"
	"compress/gzip"
	"errors"
	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
	"github.com/zohu/zgin/zbuff"
	"github.com/zohu/zgin/zlog"
	"io"
	"sync"
)

var (
	ErrClosed         = errors.New("connect is closed")    // 连接关闭
	ErrSendBufferFull = errors.New("message buff is full") // 待发送消息缓冲池满
	ErrHomeFull       = errors.New("home has reached its maximum")
)

type Websocket struct {
	isConnected bool
	Conn        *websocket.Conn
	rmt         *sync.RWMutex
	onConnected func()
	onMessage   func(msg []byte)
	onErr       func(err error)
}

type Message []byte

func NewMessage(msg []byte) Message {
	return msg
}
func NewMessageMarshal(data any) Message {
	d, _ := sonic.Marshal(data)
	return d
}
func (m Message) Unmarshal(data any) error {
	return sonic.Unmarshal(m, data)
}
func (m Message) Bytes() []byte {
	return m
}
func (m Message) Gzip() Message {
	buff := zbuff.New()
	defer buff.Free()
	g := gzip.NewWriter(buff)
	defer g.Close()
	_, _ = g.Write(m)
	return buff.Bytes()
}
func (m Message) UnGzip() Message {
	r := bytes.NewReader(m)
	g, err := gzip.NewReader(r)
	if err != nil {
		zlog.Warnf("gzip.NewReader err: %v", err)
	}
	defer g.Close()
	b, _ := io.ReadAll(g)
	return b
}
