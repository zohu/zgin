package zws

import (
	"github.com/panjf2000/ants/v2"
	"github.com/zohu/zgin/zmap"
	"math"
	"sync"
)

type WebsocketHome[T any] interface {
	Add(ID string, s WebsocketServer[T]) error
	Load(ID string) (WebsocketServer[T], bool)
	LoadFunc(func(T) bool) (WebsocketServer[T], bool)
	Remove(ID string)
	Broadcast(msg *Message)
	BroadcastWithFilter(msg *Message, filter func(ID string, data T) bool)
	OnlineSize() int
}

type Home[T any] struct {
	opts   *Options
	serves zmap.ConcurrentMap[string, WebsocketServer[T]]
	pool   *ants.Pool
}

func NewHome[T any](opts ...*Options) *Home[T] {
	h := new(Home[T])
	h.serves = zmap.New[WebsocketServer[T]]()
	if len(opts) > 0 {
		h.opts = opts[0]
	}
	h.opts.Validate()
	return h
}
func (h *Home[T]) Add(ID string, s WebsocketServer[T]) error {
	if h.serves.Count() > h.opts.HomeMaxSize {
		return ErrHomeFull
	}
	h.serves.Set(ID, s)
	return nil
}
func (h *Home[T]) Load(ID string) (WebsocketServer[T], bool) {
	return h.serves.Get(ID)
}
func (h *Home[T]) LoadFunc(fn func(T) bool) (WebsocketServer[T], bool) {
	for s := range h.serves.Iter() {
		if fn(s.Val.GetData()) {
			return s.Val, true
		}
	}
	return nil, false
}
func (h *Home[T]) Remove(ID string) {
	if s, ok := h.Load(ID); ok {
		s.Release()
		h.serves.Remove(ID)
	}
}
func (h *Home[T]) Broadcast(msg *Message) {
	if h.serves.Count() == 0 {
		return
	}
	size := int(math.Min(float64(h.serves.Count()), float64(h.opts.HomeBroadcastPoolMaxSize)))
	pool, _ := ants.NewPool(size)
	var wg sync.WaitGroup
	h.serves.IterCb(func(ID string, s WebsocketServer[T]) {
		wg.Add(1)
		_ = pool.Submit(func() {
			_ = s.Send(msg)
			wg.Done()
		})
	})
	wg.Wait()
}
func (h *Home[T]) BroadcastWithFilter(msg *Message, filter func(ID string, data T) bool) {
	if h.serves.Count() == 0 {
		return
	}
	size := int(math.Min(float64(h.serves.Count()), float64(h.opts.HomeBroadcastPoolMaxSize)))
	pool, _ := ants.NewPool(size)
	var wg sync.WaitGroup
	h.serves.IterCb(func(ID string, c WebsocketServer[T]) {
		if ok := filter(ID, c.GetData()); ok {
			wg.Add(1)
			_ = pool.Submit(func() {
				_ = c.Send(msg)
				wg.Done()
			})
		}
	})
	wg.Wait()
}
func (h *Home[T]) OnlineSize() int {
	return h.serves.Count()
}
