package logs

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func TestRingBufferAppendUnderCapacity(t *testing.T) {
	r := newRingBuffer(5)
	r.append(services.LogEvent{Content: "line1"})
	r.append(services.LogEvent{Content: "line2"})

	if r.len() != 2 {
		t.Fatalf("len() = %d, want 2", r.len())
	}
	all := r.all()
	if all[0].Content != "line1" || all[1].Content != "line2" {
		t.Fatalf("all() = %+v, want [line1 line2]", all)
	}
}

func TestRingBufferEvictsOldestOverCapacity(t *testing.T) {
	r := newRingBuffer(3)
	for i := 0; i < 5; i++ {
		r.append(services.LogEvent{Content: string(rune('a' + i))})
	}

	if r.len() != 3 {
		t.Fatalf("len() = %d, want 3", r.len())
	}
	all := r.all()
	want := []string{"c", "d", "e"}
	for i, w := range want {
		if all[i].Content != w {
			t.Errorf("all()[%d] = %q, want %q", i, all[i].Content, w)
		}
	}
}

func TestRingBufferCapacityClampedToAtLeastOne(t *testing.T) {
	r := newRingBuffer(0)
	r.append(services.LogEvent{Content: "only"})
	r.append(services.LogEvent{Content: "second"})

	if r.len() != 1 {
		t.Fatalf("len() = %d, want 1", r.len())
	}
	if r.all()[0].Content != "second" {
		t.Fatalf("all()[0].Content = %q, want %q", r.all()[0].Content, "second")
	}
}
