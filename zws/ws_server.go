package zws

import (
	"github.com/gorilla/websocket"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"net/http"
	"sync"
	"time"
)

type WebsocketServer[T any] interface {
	Start() error
	Send(msg *Message) error
	OnConnected(f func())
	OnMessage(f func(msg *Message))
	OnErr(func(err error))
	IsClose() bool
	Release()
	SetData(data T)
	GetData() T
	UnsetData()
}
type Serve[T any] struct {
	Websocket
	opts *Options
	data T // 每个服务端的临时数据
	buff chan *Message
}

func NewServe[T any](w http.ResponseWriter, r *http.Request, h http.Header, opts ...*Options) (WebsocketServer[T], error) {
	s := new(Serve[T])
	s.rmt = new(sync.RWMutex)
	if len(opts) > 0 {
		s.opts = opts[0]
	}
	s.opts.Validate()
	upgrader := websocket.Upgrader{}
	if s.opts.ServeAllowAllOrigin {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	} else if s.opts.ServeCheckOrigin != nil {
		upgrader.CheckOrigin = s.opts.ServeCheckOrigin
	}
	var err error
	s.Conn, err = upgrader.Upgrade(w, r, h)
	if err != nil {
		return nil, err
	}
	s.isConnected = true
	s.buff = make(chan *Message, s.opts.MessageBufferSize)
	return s, nil
}
func (s *Serve[T]) Send(msg *Message) error {
	if s.IsClose() {
		return ErrClosed
	}
	select {
	case s.buff <- msg:
		return nil
	default:
		return ErrSendBufferFull
	}
}
func (s *Serve[T]) OnConnected(f func()) {
	s.onConnected = f
}
func (s *Serve[T]) OnMessage(f func(msg *Message)) {
	s.onMessage = f
}
func (s *Serve[T]) OnErr(f func(err error)) {
	s.onErr = f
}
func (s *Serve[T]) IsClose() bool {
	s.rmt.RLock()
	defer s.rmt.RUnlock()
	return !s.isConnected
}
func (s *Serve[T]) Release() {
	if s.IsClose() {
		return
	}
	s.rmt.Lock()
	s.isConnected = false
	_ = s.Conn.Close()
	close(s.buff)
	s.rmt.Unlock()
}
func (s *Serve[T]) SetData(data T) {
	s.data = data
}
func (s *Serve[T]) GetData() T {
	return s.data
}
func (s *Serve[T]) UnsetData() {
	s.data = *new(T)
}
func (s *Serve[T]) Start() error {
	var exit = make(chan error, 1)
	go s.read(exit)
	go s.send(exit)
	if s.onConnected != nil {
		go s.onConnected()
	}
	select {
	case err := <-exit:
		s.Release()
		return err
	}
}
func (s *Serve[T]) read(exit chan error) {
	for {
		t, d, err := s.Conn.ReadMessage()
		if err != nil {
			if s.onErr != nil {
				go s.onErr(err)
			}
			exit <- err
			return
		}
		switch t {
		case websocket.PingMessage:
			_ = s.Send(NewMessage())
		case websocket.CloseMessage:
			if s.onErr != nil {
				go s.onErr(ErrClosed)
			}
			exit <- ErrClosed
			return
		case websocket.BinaryMessage:
			msg := NewMessage()
			if s.opts.Gzip {
				msg.MsgUnGzip(d)
			} else {
				msg.WithBinary(d)
			}
			if msg.Event() == MessagePing {
				_ = s.Send(NewMessage())
				continue
			}
			if s.onMessage != nil {
				go s.onMessage(msg)
			}
		case websocket.TextMessage:
			data := string(d)
			if len(data) >= 5 {
				msg := NewMessage().WithString(data[5:]).WithEvent(MessageCode(data[:4]))
				if msg.Event() == MessagePing {
					_ = s.Send(NewMessage())
				} else {
					go s.onMessage(msg)
				}
			}
		default:
			zlog.Warnf("websocket: unknown message type %d", t)
		}
	}
}
func (s *Serve[T]) send(exit chan error) {
	for {
		select {
		case msg := <-s.buff:
			if s.IsClose() {
				return
			}
			s.rmt.Lock()
			_ = s.Conn.SetWriteDeadline(time.Now().Add(s.opts.WriteWait))
			d := zutil.When(s.opts.Gzip, msg.MsgGzip(), msg.MsgBytes())
			err := s.Conn.WriteMessage(websocket.BinaryMessage, d)
			s.rmt.Unlock()
			if err != nil {
				exit <- err
				return
			}
		}
	}
}
