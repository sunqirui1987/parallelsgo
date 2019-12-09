package parallelsgo

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type ABC struct {
	CBA int
}

func TestNormal(t *testing.T) {
	var (
		abcs = make(map[int]*ABC)
		g    Parallels
		err  error
	)
	for i := 0; i < 10; i++ {
		abcs[i] = &ABC{CBA: i}
	}
	g.Go(func(ctx context.Context, args interface{}) (err error) {
		i := args.(int)
		abcs[i].CBA++
		return
	}, 1)
	g.Go(func(ctx context.Context, args interface{}) (err error) {
		i := args.(int)
		abcs[i].CBA++
		return
	}, 2)
	if err = g.Wait(); err != nil {
		t.Log(err)
	}
	t.Log(abcs)
}

func TestWithCancel(t *testing.T) {
	g := WithCancel(context.Background())
	g.Go(func(ctx context.Context, args interface{}) error {
		time.Sleep(100 * time.Millisecond)
		return fmt.Errorf("boom")
	}, nil)
	var doneErr error
	g.Go(func(ctx context.Context, args interface{}) error {
		select {
		case <-ctx.Done():
			doneErr = ctx.Err()
		}
		return doneErr
	}, nil)
	g.Wait()
	if doneErr != context.Canceled {
		t.Error("error should be Canceled")
	}
}
