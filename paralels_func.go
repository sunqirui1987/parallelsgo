package parallelsgo

import (
	"context"
)

type Paralels_Func struct {
	PFunc func(ctx context.Context, args interface{}) error
	Args  interface{}
}

func NewParalels_Func(f func(ctx context.Context, args interface{}) error, args interface{}) *Paralels_Func {
	return &Paralels_Func{
		PFunc: f,
		Args:  args,
	}
}
