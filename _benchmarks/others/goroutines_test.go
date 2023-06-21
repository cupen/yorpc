package yorpc

import (
	"sync"
	"testing"
)

func BenchmarkGoroutineSwitch(b *testing.B) {
	ch := make(chan struct{}, 1_000_000)
	ch2 := make(chan struct{}, 1_000_000)
	ch3 := make(chan struct{}, 1_000_000)
	b.Cleanup(func() {
		close(ch)
		close(ch2)
		close(ch3)
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ch:
				ch <- struct{}{}
			case <-ch2:
				ch2 <- struct{}{}
				return
			case <-ch3:
				ch3 <- struct{}{}
				return
			}
		}
	}()

	for i := 0; i < b.N; i++ {
		ch <- struct{}{}
		<-ch
	}
	ch2 <- struct{}{}
	<-ch2
	wg.Wait()
}
