package zws

import (
	"github.com/panjf2000/ants/v2"
	"github.com/zohu/zgin/zmap"
	"math"
	"sync"
)

type Home struct {
	opts   *Options
	serves zmap.ConcurrentMap[string, *Serve]
	pool   *ants.Pool
}

func NewHome(opts ...*Options) *Home {
	h := new(Home)
	h.serves = zmap.New[*Serve]()
	if len(opts) > 0 {
		h.opts = opts[0]
	}
	h.opts.Validate()
	return h
}
func (h *Home) Add(ID string, s *Serve) error {
	if h.serves.Count() > h.opts.HomeMaxSize {
		return ErrHomeFull
	}
	h.serves.Set(ID, s)
	return nil
}
func (h *Home) Load(ID string) (*Serve, bool) {
	return h.serves.Get(ID)
}
func (h *Home) Remove(ID string) {
	if s, ok := h.Load(ID); ok {
		s.Release()
		h.serves.Remove(ID)
	}
}
func (h *Home) Broadcast(msg []byte) {
	if h.serves.Count() == 0 {
		return
	}
	size := int(math.Min(float64(h.serves.Count()), float64(h.opts.HomeBroadcastPoolMaxSize)))
	pool, _ := ants.NewPool(size)
	var wg sync.WaitGroup
	h.serves.IterCb(func(ID string, s *Serve) {
		wg.Add(1)
		_ = pool.Submit(func() {
			_ = s.Send(msg)
			wg.Done()
		})
	})
	wg.Wait()
}
func (h *Home) BroadcastWithFilter(msg []byte, filter func(s *Serve) bool) {
	if h.serves.Count() == 0 {
		return
	}
	size := int(math.Min(float64(h.serves.Count()), float64(h.opts.HomeBroadcastPoolMaxSize)))
	pool, _ := ants.NewPool(size)
	var wg sync.WaitGroup
	h.serves.IterCb(func(ID string, s *Serve) {
		if ok := filter(s); ok {
			wg.Add(1)
			_ = pool.Submit(func() {
				_ = s.Send(msg)
				wg.Done()
			})
		}
	})
	wg.Wait()
}
func (h *Home) OnlineSize() int {
	return h.serves.Count()
}
