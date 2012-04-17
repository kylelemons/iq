package iq

type Ring struct {
	cnt, i int
	data []Type
}

func (rb *Ring) Empty() bool {
	return rb.cnt == 0
}

func (rb *Ring) Peek() Type {
	return rb.data[rb.i]
}

func (rb *Ring) Enqueue(x Type) {
	if rb.cnt >= len(rb.data) {
		rb.grow(2 * rb.cnt + 1)
	}
	rb.data[(rb.i + rb.cnt) % len(rb.data)] = x
	rb.cnt++
}

func (rb *Ring) Dequeue() {
	rb.cnt, rb.i = rb.cnt - 1, (rb.i + 1) % len(rb.data)
}

func (rb *Ring) grow(newSize int) {
	newData := make([]Type, newSize)

	n := copy(newData, rb.data[rb.i:])
	copy(newData[n:], rb.data[:rb.cnt-n])

	rb.i = 0
	rb.data = newData
}

func RingIQ(in <-chan Type, next chan<- Type) {
	var rb Ring
	defer func() {
		for !rb.Empty() {
			next <- rb.Peek()
			rb.Dequeue()
		}
		close(next)
	}()

	for {
		if rb.Empty() {
			v, ok := <-in
			if !ok {
				return
			}
			rb.Enqueue(v)
		}

		select {
		case next <- rb.Peek():
			rb.Dequeue()
		case v, ok := <-in:
			if !ok {
				return
			}
			rb.Enqueue(v)
		}
	}
}
