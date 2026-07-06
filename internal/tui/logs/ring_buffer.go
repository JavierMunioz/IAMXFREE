package logs

import "github.com/JavierMunioz/IAMXFREE/internal/services"

// ringBuffer retains a bounded, configurable number of received log
// events for scrollback — old events are evicted as new ones arrive, so
// keeping the logs view open indefinitely never grows memory without
// bound.
type ringBuffer struct {
	capacity int
	events   []services.LogEvent
}

func newRingBuffer(capacity int) *ringBuffer {
	if capacity < 1 {
		capacity = 1
	}
	return &ringBuffer{capacity: capacity}
}

// append records e, evicting the oldest event if the buffer is over
// capacity.
func (r *ringBuffer) append(e services.LogEvent) {
	r.events = append(r.events, e)
	if len(r.events) > r.capacity {
		r.events = r.events[len(r.events)-r.capacity:]
	}
}

// len reports how many events are currently retained.
func (r *ringBuffer) len() int { return len(r.events) }

// all returns every event currently retained, oldest first.
func (r *ringBuffer) all() []services.LogEvent { return r.events }
