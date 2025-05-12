package zch

import (
	"fmt"
	"github.com/zohu/zgin/zmap"
	"runtime"
	"strings"
	"time"
)

type Item struct {
	value      string
	expiration int64
}

func (i Item) Expired() bool {
	if i.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > i.expiration
}

type Memory struct {
	*memory
}
type memory struct {
	expiration time.Duration
	cmap       zmap.ConcurrentMap[string, Item]
	janitor    *janitor
	prefix     string
}

func (c *memory) Get(k string) (string, bool) {
	item, found := c.cmap.Get(c.key(k))
	if !found || item.Expired() {
		return "", false
	}
	return item.value, true
}

func (c *memory) GetWithExpiration(k string) (string, time.Time, bool) {
	item, found := c.cmap.Get(c.key(k))
	if !found || item.Expired() {
		return "", time.Time{}, false
	}
	return item.value, time.Unix(0, item.expiration), true
}

func (c *memory) Delete(k string) {
	c.cmap.Remove(c.key(k))
}

func (c *memory) Set(k string, x string, d time.Duration) {
	var e int64
	if d <= 0 {
		d = c.expiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.cmap.Set(c.key(k), Item{
		value:      x,
		expiration: e,
	})
}
func (c *memory) SetNX(k, v string, d time.Duration) error {
	_, ok := c.Get(k)
	if ok {
		return fmt.Errorf("key already exists")
	}
	c.Set(k, v, d)
	return nil
}
func (c *memory) Replace(k, v string, d time.Duration) error {
	_, ok := c.Get(k)
	if !ok {
		return fmt.Errorf("key does not exist")
	}
	c.Set(k, v, d)
	return nil
}
func (c *memory) Items() map[string]string {
	m := make(map[string]string)
	c.cmap.IterCb(func(key string, item Item) {
		if !item.Expired() {
			m[key] = item.value
		}
	})
	return m
}
func (c *memory) Count() int {
	c.expired()
	return c.cmap.Count()
}
func (c *memory) Flush() {
	c.cmap.Clear()
}
func (c *memory) expired() {
	c.cmap.IterCb(func(key string, item Item) {
		if item.Expired() {
			c.cmap.Remove(key)
		}
	})
}
func (c *memory) key(k string) string {
	if c.prefix != "" && !strings.HasPrefix(k, fmt.Sprintf("%s:", c.prefix)) {
		return fmt.Sprintf("%s:%s", c.prefix, k)
	}
	return k
}

type janitor struct {
	interval time.Duration
	stop     chan bool
}

func (j *janitor) Run(c *memory) {
	ticker := time.NewTicker(j.interval)
	for {
		select {
		case <-ticker.C:
			c.expired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func stopJanitor(c *Memory) {
	c.janitor.stop <- true
}

func runJanitor(c *memory, ci time.Duration) {
	j := &janitor{
		interval: ci,
		stop:     make(chan bool),
	}
	c.janitor = j
	go j.Run(c)
}

func newMemory(de time.Duration, prefix string, m zmap.ConcurrentMap[string, Item]) *memory {
	if de == 0 {
		de = time.Minute * 5
	}
	c := &memory{
		prefix:     prefix,
		expiration: de,
		cmap:       m,
	}
	if c.prefix != "" {
		c.cmap.IterCb(func(key string, item Item) {
			c.cmap.Set(c.key(key), item)
		})
	}
	return c
}

func newMemoryWithJanitor(de time.Duration, ci time.Duration, prefix string, m zmap.ConcurrentMap[string, Item]) *Memory {
	c := newMemory(de, prefix, m)
	C := &Memory{c}
	if ci > 0 {
		runJanitor(c, ci)
		runtime.SetFinalizer(C, stopJanitor)
	}
	return C
}

func NewMemory(defaultExpiration, cleanupInterval time.Duration, prefix string) *Memory {
	return newMemoryWithJanitor(defaultExpiration, cleanupInterval, prefix, zmap.New[Item]())
}

func NewMemoryFrom(defaultExpiration, cleanupInterval time.Duration, prefix string, m zmap.ConcurrentMap[string, Item]) *Memory {
	return newMemoryWithJanitor(defaultExpiration, cleanupInterval, prefix, m)
}
