package zws

import (
	"bytes"
	"compress/gzip"
	"io"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/zohu/zgin"
	"github.com/zohu/zgin/zbuff"
	"github.com/zohu/zgin/zlog"
)

type MessageCode string

func (m MessageCode) String() string {
	return string(m)
}

const (
	MessagePing MessageCode = "0000"
)

type MessageMode int

const (
	MessageModeText MessageMode = iota
	MessageModeJson
	MessageModeBinary
	MessageModeNumber
)

type Message struct {
	event MessageCode
	mode  MessageMode
	data  []byte
}

func NewMessage() *Message {
	return &Message{event: MessagePing, mode: MessageModeText}
}
func (m *Message) WithEvent(event MessageCode) *Message {
	m.event = event
	return m
}
func (m *Message) WithStruct(data interface{}) *Message {
	var err error
	m.data, err = sonic.Marshal(data)
	if err != nil {
		zlog.Errorf("Message.WithStruct Marshal err: %v", err)
	}
	m.mode = MessageModeJson
	return m
}
func (m *Message) WithString(data string) *Message {
	m.data = []byte(data)
	m.mode = MessageModeText
	return m
}
func (m *Message) WithInt(data int64) *Message {
	// 利用文本转换，不关心大小端问题
	m.data = []byte(strconv.FormatInt(data, 10))
	m.mode = MessageModeNumber
	return m
}
func (m *Message) WithFloat(data float64, prec int) *Message {
	// 利用文本转换，不关心大小端问题且精度可控
	m.data = []byte(strconv.FormatFloat(data, 'f', prec, 64))
	m.mode = MessageModeNumber
	return m
}
func (m *Message) WithBinary(data []byte) *Message {
	m.data = append([]byte{}, data...)
	m.mode = MessageModeBinary
	return m
}
func (m *Message) Event() MessageCode {
	return m.event
}
func (m *Message) Mode() MessageMode {
	return m.mode
}
func (m *Message) Map() map[string]interface{} {
	var data map[string]interface{}
	if err := sonic.Unmarshal(m.data, &data); err != nil {
		zlog.Errorf("Message.Map Unmarshal error: %v", err)
	}
	return data
}
func (m *Message) Bind(dst interface{}) error {
	if err := sonic.Unmarshal(m.data, dst); err != nil {
		return err
	}
	return zgin.Validator().Struct(dst)
}
func (m *Message) String() string {
	return string(m.data)
}
func (m *Message) Int() int64 {
	i, err := strconv.ParseInt(string(m.data), 10, 64)
	if err != nil {
		zlog.Errorf("Message.Int ParseInt error: %v", err)
	}
	return i
}
func (m *Message) MsgBytes() []byte {
	buff := zbuff.New()
	defer buff.Free()
	buff.WriteString(m.event.String())
	buff.WriteString(strconv.Itoa(int(m.mode)))
	buff.WriteString(string(m.data))
	return buff.Clone()
}
func (m *Message) MsgGzip() []byte {
	buff := zbuff.New()
	defer buff.Free()
	gz := gzip.NewWriter(buff)
	_, _ = gz.Write(m.MsgBytes())
	_ = gz.Flush()
	_ = gz.Close()
	return buff.Clone()
}
func (m *Message) MsgUnGzip(msg []byte) *Message {
	gz, err := gzip.NewReader(bytes.NewBuffer(msg))
	if err != nil {
		zlog.Errorf("Message.MsgUnGzip err: %v", err)
		return m
	}
	defer gz.Close()
	d, _ := io.ReadAll(gz)
	str := string(d)
	if len([]rune(str)) >= 4 {
		m.event = MessageCode(str[:4])
	}
	if len([]rune(str)) >= 5 {
		model, _ := strconv.Atoi(str[4:5])
		m.mode = MessageMode(model)
	}
	if len([]rune(str)) >= 6 {
		m.data = []byte(str[5:])
	}
	return m
}
