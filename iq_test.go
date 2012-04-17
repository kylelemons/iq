package iq

import (
	"testing"
	"time"
)

func TestIQ(t *testing.T) {
	vals := []Type{
		1, 2, 4, 42,
	}
	imps := map[string]func(<-chan Type, chan<- Type){
		"slice": SliceIQ,
		"ring":  RingIQ,
	}

	for imp, f := range imps {
		in, out := make(chan Type), make(chan Type)
		go f(in, out)

		for _, v := range vals {
			in <- v
		}
		close(in)

		var i int
		for v := range out {
			if i >= len(vals) {
				t.Errorf("%s: %d. extra %v", imp, v)
			} else if got, want := v, vals[i]; got != want {
				t.Errorf("%s: %d. got %v, want %v", imp, i, got, want)
			}
			i++
		}
		for j, v := range vals[i:] {
			t.Errorf("%s: %d. missing %v", imp, i+j, v)
		}
	}
}

func TestRing(t *testing.T) {
	bench := func(n int, name string, f func(<-chan Type, chan<- Type), batch, offset int) {
		timeout := time.After(1*time.Second)
		send, recv := make(chan Type), make(chan Type)
		go f(send, recv)
		for i := 0; i < offset; i++ {
			send <- Type(i)
		}
		for i := 0; i < n-offset; i++ {
			send <- Type(offset+i)

			// Drain the queue
			if i % batch == batch - 1 {
				for bi := 0; bi < batch; bi++ {
					var v Type
					select {
					case v = <-recv:
					case <-timeout:
						t.Errorf("%s: %d (drain %d). timeout", name, i, batch)
						return
					}
					if got, want := v, Type(i+1-offset-batch+bi); got != want {
						t.Errorf("%s: %d (drain %d). got %d, want %d", name, i, batch, got, want)
					}
				}
			}
		}
		close(send)
		
		i := n-offset
		for v := range recv {
			if got, want := v, Type(i); got != want {
				t.Errorf("%d (drain %d). got %d, want %d", i, batch, got, want)
			}
			i++
		}
	}

	bench(100, "slice", SliceIQ, 10, 0)
	bench(100, "ring", RingIQ, 10, 0)
}

func bench(n int, f func(<-chan Type, chan<- Type), batch, offset int) {
	send, recv := make(chan Type), make(chan Type)
	go f(send, recv)
	for i := 0; i < offset; i++ {
		send <- Type(i)
	}
	for i := 0; i < n-offset; i++ {
		send <- Type(offset+i)

		// Drain the queue
		if i % batch == batch - 1 {
			for bi := 0; bi < batch; bi++ {
				<-recv
			}
		}
	}
	close(send)
	for _ = range recv {}
}

func Benchmark_SliceIQ___1___0(b *testing.B) { bench(b.N, SliceIQ, 1, 0) }
func Benchmark__RingIQ___1___0(b *testing.B) { bench(b.N, RingIQ, 1, 0) }

func Benchmark_SliceIQ__10___0(b *testing.B) { bench(b.N, SliceIQ, 10, 0) }
func Benchmark__RingIQ__10___0(b *testing.B) { bench(b.N, RingIQ, 10, 0) }

func Benchmark_SliceIQ__10___5(b *testing.B) { bench(b.N, SliceIQ, 10, 5) }
func Benchmark__RingIQ__10___5(b *testing.B) { bench(b.N, RingIQ, 10, 5) }

func Benchmark_SliceIQ_100___0(b *testing.B) { bench(b.N, SliceIQ, 100, 0) }
func Benchmark__RingIQ_100___0(b *testing.B) { bench(b.N, RingIQ, 100, 0) }

func Benchmark_SliceIQ_100__50(b *testing.B) { bench(b.N, SliceIQ, 100, 50) }
func Benchmark__RingIQ_100__50(b *testing.B) { bench(b.N, RingIQ, 100, 50) }
