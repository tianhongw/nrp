package util

import "sync"

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (w *WaitGroupWrapper) Warp(f func()) {
	w.Add(1)

	go func() {
		defer w.Done()
		f()
	}()
}
