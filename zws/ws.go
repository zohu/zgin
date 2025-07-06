package zws

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

/**
封装长连接，约定如下：
- 客户端实现ping，服务端实现pong
- 消息自动转换为统一的gzip压缩二进制；
- 客户端实现自动重连；
- 服务端实现多端管理和广播；
- 交互消息统一格式为：消息类型(4位)+数据格式(1位)+数据
*/

var (
	ErrClosed         = errors.New("连接关闭")
	ErrSendBufferFull = errors.New("待发送消息缓冲池满")
	ErrHomeFull       = errors.New("群组超限")
)

type Websocket struct {
	isConnected bool
	Conn        *websocket.Conn
	rmt         *sync.RWMutex
	onConnected func()
	onMessage   func(msg *Message)
	onErr       func(err error)
}
