package yorpc

import "testing"

func BenchmarkGoroutineSwitch(b *testing.B) {
	ch := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})
	b.Cleanup(func() {
		close(ch)
		close(ch2)
		close(ch3)
	})

	go func() {
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
}
