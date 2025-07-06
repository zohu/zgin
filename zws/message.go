package zws

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/zohu/zgin/zbuff"
	"io"
	"strconv"
)

type MessageType string

func (m MessageType) String() string {
	return string(m)
}

const (
	MessagePing MessageType = "0000"
)

type MessageMode int

const (
	MessageModeText MessageMode = iota
	MessageModeJson
	MessageModeBinary
)

type Message struct {
	event MessageType
	mode  MessageMode
	data  string
}

func NewMessage() *Message {
	return &Message{event: MessagePing}
}
func (m *Message) WithEvent(event MessageType) *Message {
	m.event = event
	return m
}
func (m *Message) WithStruct(data interface{}) *Message {
	m.data, _ = sonic.MarshalString(data)
	m.mode = MessageModeJson
	return m
}
func (m *Message) WithString(data string) *Message {
	m.data = data
	m.mode = MessageModeText
	return m
}
func (m *Message) WithInt(data int64) *Message {
	m.data = strconv.FormatInt(data, 10)
	m.mode = MessageModeText
	return m
}
func (m *Message) WithBinary(data []byte) *Message {
	m.data = string(data)
	m.mode = MessageModeBinary
	return m
}
func (m *Message) Event() MessageType {
	return m.event
}
func (m *Message) Mode() MessageMode {
	return m.mode
}
func (m *Message) Map() map[string]interface{} {
	var data map[string]interface{}
	_ = sonic.UnmarshalString(m.data, &data)
	return data
}
func (m *Message) Bind(dst interface{}) error {
	return sonic.UnmarshalString(m.data, &dst)
}
func (m *Message) String() string {
	return m.data
}
func (m *Message) Int() int64 {
	i, _ := strconv.ParseInt(m.data, 10, 64)
	return i
}
func (m *Message) MsgBytes() []byte {
	return []byte(fmt.Sprintf("%s%d%s", m.event, m.mode, m.data))
}
func (m *Message) MsgGzip() []byte {
	buff := zbuff.New()
	defer buff.Free()
	gz := gzip.NewWriter(buff)
	_, _ = gz.Write(m.MsgBytes())
	_ = gz.Flush()
	_ = gz.Close()
	return buff.Bytes()
}
func (m *Message) MsgUnGzip(msg []byte) *Message {
	gz, _ := gzip.NewReader(bytes.NewBuffer(msg))
	defer gz.Close()
	d, _ := io.ReadAll(gz)
	str := string(d)
	if len([]rune(str)) >= 4 {
		m.event = MessageType(str[:4])
	}
	if len([]rune(str)) >= 5 {
		model, _ := strconv.Atoi(str[4:5])
		m.mode = MessageMode(model)
	}
	if len([]rune(str)) >= 6 {
		m.data = str[5:]
	}
	return m
}
