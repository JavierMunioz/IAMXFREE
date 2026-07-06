package runtimehost

import "sync"

// defaultOutputBufferCapacity bounds how many output lines a single
// process's outputBuffer retains — old lines are evicted as new ones
// arrive, so a long-running process's captured output never grows without
// bound in memory.
const defaultOutputBufferCapacity = 2000

// outputBuffer captures a bounded backlog of a process's stdout/stderr and
// lets any number of OutputStreams subscribe to it: each subscriber first
// replays whatever backlog is currently retained, then receives new chunks
// as they're appended, until finish is called.
type outputBuffer struct {
	mu       sync.Mutex
	cond     *sync.Cond
	capacity int
	chunks   []OutputChunk
	baseSeq  int64
	done     bool
	err      error
}

func newOutputBuffer(capacity int) *outputBuffer {
	b := &outputBuffer{capacity: capacity}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// append records c, evicting the oldest chunk(s) if the buffer is over
// capacity, and wakes any subscriber currently waiting for more data.
func (b *outputBuffer) append(c OutputChunk) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.done {
		return
	}
	b.chunks = append(b.chunks, c)
	if len(b.chunks) > b.capacity {
		excess := len(b.chunks) - b.capacity
		b.chunks = b.chunks[excess:]
		b.baseSeq += int64(excess)
	}
	b.cond.Broadcast()
}

// finish marks the buffer as complete (the process exited), with err set
// only if capture ended abnormally. Every subscriber still waiting wakes up
// and observes it via Err() once it has drained the backlog.
func (b *outputBuffer) finish(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.done {
		return
	}
	b.done = true
	b.err = err
	b.cond.Broadcast()
}

// subscribe returns a new OutputStream over b, starting from the oldest
// chunk currently retained.
func (b *outputBuffer) subscribe() *bufferStream {
	s := &bufferStream{
		buffer: b,
		out:    make(chan OutputChunk),
		closed: make(chan struct{}),
	}
	go s.run()
	return s
}

// bufferStream is outputBuffer's OutputStream implementation.
type bufferStream struct {
	buffer    *outputBuffer
	out       chan OutputChunk
	closed    chan struct{}
	closeOnce sync.Once
	err       error
}

var _ OutputStream = (*bufferStream)(nil)

func (s *bufferStream) run() {
	defer close(s.out)

	b := s.buffer
	var seq int64

	b.mu.Lock()
	for {
		select {
		case <-s.closed:
			b.mu.Unlock()
			return
		default:
		}

		start := seq - b.baseSeq
		if start < 0 {
			start = 0
		}
		if int(start) < len(b.chunks) {
			pending := append([]OutputChunk(nil), b.chunks[start:]...)
			seq = b.baseSeq + int64(len(b.chunks))
			b.mu.Unlock()
			for _, c := range pending {
				select {
				case s.out <- c:
				case <-s.closed:
					return
				}
			}
			b.mu.Lock()
			continue
		}

		if b.done {
			s.err = b.err
			b.mu.Unlock()
			return
		}

		// Wait releases b.mu and reacquires it before returning, either on
		// append/finish's Broadcast or on Close's Broadcast below.
		b.cond.Wait()
	}
}

func (s *bufferStream) Chunks() <-chan OutputChunk { return s.out }

func (s *bufferStream) Err() error { return s.err }

func (s *bufferStream) Close() error {
	s.closeOnce.Do(func() {
		close(s.closed)
		s.buffer.mu.Lock()
		s.buffer.cond.Broadcast()
		s.buffer.mu.Unlock()
	})
	return nil
}
