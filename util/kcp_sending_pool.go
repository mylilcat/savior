package util

import (
	"log"
	"runtime"
	"sync"
)

var pool *sync.Pool

type kcpSender struct {
	p *sync.Pool
	t chan func()
}

func (ks *kcpSender) run() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 1024)
				n := runtime.Stack(buf, false)
				log.Printf("Recovered from panic: %v\nStack trace:\n%s", r, buf[:n])
			}
		}()
		for f := range ks.t {
			if f == nil {
				return
			}
			f()
			ks.p.Put(ks)
			return
		}
	}()
}

func KcpSendPoolInit() {
	pool = &sync.Pool{}
	pool.New = func() any {
		return &kcpSender{
			p: pool,
			t: make(chan func()),
		}
	}
}

func KcpSend(f func()) {
	ks := pool.Get().(*kcpSender)
	ks.run()
	ks.t <- f
}
