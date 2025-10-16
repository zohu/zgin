package zws

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/gorilla/websocket"
	"github.com/zohu/zgin/zutil"
	"github.com/zohu/zlog"
)

type WebsocketClient interface {
	Start(ctx context.Context) error
	Send(msg *Message) error
	OnConnected(f func())
	OnMessage(f func(msg *Message))
	OnErr(f func(err error))
	Release()
}
type Client struct {
	Websocket
	addr string
	opts *Options
	buff chan *Message
}

func NewClient(addr string, opts ...*Options) WebsocketClient {
	o := &Options{}
	if len(opts) > 0 {
		o = opts[0]
	}
	o.Validate()
	return &Client{
		addr: addr,
		opts: o,
		buff: make(chan *Message, o.MessageBufferSize),
	}
}

func (c *Client) Send(msg *Message) error {
	if c.IsClose() {
		return ErrClosed
	}
	select {
	case c.buff <- msg:
		return nil
	default:
		return ErrSendBufferFull
	}
}
func (c *Client) OnConnected(f func()) {
	c.onConnected = f
}
func (c *Client) OnMessage(f func(msg *Message)) {
	c.onMessage = f
}
func (c *Client) OnErr(f func(err error)) {
	c.onErr = f
}
func (c *Client) IsClose() bool {
	c.rmt.Lock()
	defer c.rmt.Unlock()
	return !c.isConnected
}
func (c *Client) Release() {
	if c.IsClose() {
		return
	}
	c.rmt.Lock()
	c.isConnected = false
	_ = c.Conn.Close()
	c.rmt.Unlock()
}
func (c *Client) Start(ctx context.Context) error {
	conn, err := backoff.Retry[*websocket.Conn](
		ctx,
		c.operation,
		backoff.WithMaxTries(c.opts.BackoffTimes),
		backoff.WithNotify(func(err error, d time.Duration) {
			zlog.Warnf("retry connect to %s after %s: %v", c.addr, d, err)
		}),
	)
	if err != nil {
		return fmt.Errorf("connect to %s err: %v", c.addr, err)
	}
	c.rmt.Lock()
	c.Conn = conn
	c.isConnected = true
	c.rmt.Unlock()
	c.Conn.SetReadLimit(c.opts.MaxMessageSize)
	var exit = make(chan error, 1)
	go c.read(exit)
	go c.send(exit)
	go c.ping(exit)

	if c.onConnected != nil {
		go c.onConnected()
	}

	select {
	case err = <-exit:
		c.Release()
		return err
	}
}

func (c *Client) operation() (*websocket.Conn, error) {
	if conn, _, err := websocket.DefaultDialer.Dial(c.addr, nil); err != nil {
		if c.onErr != nil {
			go c.onErr(err)
		}
		return nil, backoff.RetryAfter(int(c.opts.BackoffDuration.Seconds()))
	} else {
		zlog.Infof("websocket success: %s", c.addr)
		return conn, nil
	}
}
func (c *Client) ping(exit chan error) {
	t := time.NewTicker(time.Second * 45)
	for {
		select {
		case <-t.C:
			if c.IsClose() {
				t.Stop()
				exit <- ErrClosed
				return
			}
			c.buff <- NewMessage()
		}
	}
}
func (c *Client) send(exit chan error) {
	for {
		select {
		case msg := <-c.buff:
			if c.IsClose() {
				exit <- ErrClosed
				return
			}
			if msg != nil {
				c.rmt.Lock()
				_ = c.Conn.SetWriteDeadline(time.Now().Add(c.opts.WriteWait))
				d := zutil.When(c.opts.Gzip, msg.MsgGzip(), msg.MsgBytes())
				err := c.Conn.WriteMessage(websocket.BinaryMessage, d)
				c.rmt.Unlock()
				if err != nil {
					exit <- err
					return
				}
			}
		}
	}
}
func (c *Client) read(exit chan error) {
	for {
		t, d, err := c.Conn.ReadMessage()
		if err != nil {
			if c.onErr != nil {
				go c.onErr(err)
			}
			exit <- err
			return
		}
		switch t {
		case websocket.CloseMessage:
			if c.onErr != nil {
				go c.onErr(ErrClosed)
			}
			exit <- ErrClosed
			return
		case websocket.BinaryMessage:
			msg := NewMessage()
			if c.opts.Gzip {
				msg.MsgUnGzip(d)
			} else {
				msg.WithBinary(d)
			}
			if c.onMessage != nil {
				go c.onMessage(msg)
			}
		case websocket.TextMessage:
			data := string(d)
			if len(data) >= 5 {
				msg := NewMessage().WithString(data[5:]).WithEvent(MessageCode(data[:4]))
				if c.onMessage != nil {
					go c.onMessage(msg)
				}
			}
		default:
			zlog.Warnf("unsupported message type: %d", t)
		}
	}
}
