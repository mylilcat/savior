package timer

import (
	"testing"
	"time"
)

func TestName(t *testing.T) {
	timer := NewTimer(20, time.Millisecond, 1000)
	timer.Start()
	println("add:", time.Now().UnixMilli())
	timer.AddTask(func() {
		println("now:", time.Now().UnixMilli())
	}, 180)
}
