package zws

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type Serve struct {
	Websocket
	opts *Options
	data []byte // 每个服务端的临时数据
	buff chan Message
}

func NewServe(w http.ResponseWriter, r *http.Request, h http.Header, opts ...*Options) (*Serve, error) {
	s := new(Serve)
	s.rmt = new(sync.RWMutex)
	if len(opts) > 0 {
		s.opts = opts[0]
	}
	s.opts.Validate()
	upgrader := websocket.Upgrader{}
	if s.opts.ServeAllowAllOrigin {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	} else if s.opts.ServeCheckOrigin != nil {
		upgrader.CheckOrigin = s.opts.ServeCheckOrigin
	}
	var err error
	s.Conn, err = upgrader.Upgrade(w, r, h)
	if err != nil {
		return nil, err
	}
	s.isConnected = true
	s.buff = make(chan Message, s.opts.MessageBufferSize)
	return s, nil
}
func (s *Serve) OnConnected(f func()) {
	s.onConnected = f
}
func (s *Serve) OnMessage(f func(msg []byte)) {
	s.onMessage = f
}
func (s *Serve) OnErr(f func(err error)) {
	s.onErr = f
}
func (s *Serve) IsClose() bool {
	s.rmt.RLock()
	defer s.rmt.RUnlock()
	return !s.isConnected
}
func (s *Serve) Release() {
	if s.IsClose() {
		return
	}
	s.rmt.Lock()
	s.isConnected = false
	_ = s.Conn.Close()
	close(s.buff)
	s.rmt.Unlock()
}
func (s *Serve) SetData(data []byte) {
	s.data = data
}
func (s *Serve) GetData() []byte {
	return s.data
}
func (s *Serve) UnsetData() {
	s.data = nil
}
func (s *Serve) Send(msg Message) error {
	if s.IsClose() {
		return ErrClosed
	}
	if len(msg) == 0 {
		return nil
	}
	select {
	case s.buff <- msg:
		return nil
	default:
		return ErrSendBufferFull
	}
}
func (s *Serve) Start() error {
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
func (s *Serve) read(exit chan error) {
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
			_ = s.Conn.WriteMessage(websocket.PongMessage, d)
		case websocket.CloseMessage:
			if s.onErr != nil {
				go s.onErr(ErrClosed)
			}
			exit <- ErrClosed
			return
		case websocket.BinaryMessage:
			if s.onMessage != nil {
				if s.opts.Gzip {
					d = NewMessage(d).UnGzip()
				}
				go s.onMessage(d)
			}
		}
	}
}
func (s *Serve) send(exit chan error) {
	for {
		select {
		case msg := <-s.buff:
			if s.IsClose() {
				return
			}
			s.rmt.Lock()
			_ = s.Conn.SetWriteDeadline(time.Now().Add(s.opts.WriteWait))
			if s.opts.Gzip {
				msg = msg.Gzip()
			}
			err := s.Conn.WriteMessage(websocket.BinaryMessage, msg)
			s.rmt.Unlock()
			if err != nil {
				exit <- err
				return
			}
		}
	}
}
