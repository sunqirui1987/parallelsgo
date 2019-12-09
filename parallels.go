package parallelsgo

import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

// A Parallels is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Parallels is valid and does not cancel on error.
type Parallels struct {
	err     error
	wg      sync.WaitGroup
	errOnce sync.Once

	workerOnce sync.Once
	ch         chan *Paralels_Func // func(ctx context.Context) error

	chs []*Paralels_Func //func(ctx context.Context) error

	ctx    context.Context
	cancel func()
}

// WithContext create a Parallels.
// given function from Go will receive this context,
func WithContext(ctx context.Context) *Parallels {
	return &Parallels{ctx: ctx}
}

// WithCancel create a new Parallels and an associated Context derived from ctx.
//
// given function from Go will receive context derived from this ctx,
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
func WithCancel(ctx context.Context) *Parallels {
	ctx, cancel := context.WithCancel(ctx)
	return &Parallels{ctx: ctx, cancel: cancel}
}

func (g *Parallels) do(pf *Paralels_Func) {
	ctx := g.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	var err error
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 64<<10)
			buf = buf[:runtime.Stack(buf, false)]
			err = fmt.Errorf("errParallels: panic recovered: %s\n%s", r, buf)
		}
		if err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
		g.wg.Done()
	}()
	err = pf.PFunc(ctx, pf.Args)
}

// GOMAXPROCS set max goroutine to work.
func (g *Parallels) GOMAXPROCS(n int) {
	if n <= 0 {
		panic("errParallels: GOMAXPROCS must great than 0")
	}
	g.workerOnce.Do(func() {
		g.ch = make(chan *Paralels_Func, n)
		for i := 0; i < n; i++ {
			go func() {
				for pf := range g.ch {
					g.do(pf)
				}
			}()
		}
	})
}

func (g *Parallels) Go2(pf *Paralels_Func) {
	g.wg.Add(1)
	if g.ch != nil {
		select {
		case g.ch <- pf:
		default:
			g.chs = append(g.chs, pf)
		}
		return
	}
	go g.do(pf)
}

func (g *Parallels) Go(f func(ctx context.Context, args interface{}) error, args interface{}) {

	pf := NewParalels_Func(f, args)
	g.wg.Add(1)
	if g.ch != nil {
		select {
		case g.ch <- pf:
		default:
			g.chs = append(g.chs, pf)
		}
		return
	}
	go g.do(pf)
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *Parallels) Wait() error {
	if g.ch != nil {
		for _, pf := range g.chs {
			g.ch <- pf
		}
	}
	g.wg.Wait()
	if g.ch != nil {
		close(g.ch) // let all receiver exit
	}
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}
