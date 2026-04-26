package or

import (
	"testing"
	"time"
)

func TestNoChannels(t *testing.T) {
	if ch := Or(); ch != nil {
		t.Fatal("expected nil for zero channels")
	}
}

func TestSingleChannel(t *testing.T) {
	openChan := makeOpenChan()
	out := Or(openChan)
	if out != openChan {
		t.Fatal("expected the same channel to be returned for a single input")
	}
}

func TestAlreadyClosed(t *testing.T) {
	done := Or(makeOpenChan(), makeClosedChan(), makeOpenChan())
	select {
	case <-done:
	// Should be done instantly because we used closed chan
	case <-time.After(time.Millisecond):
		t.Fatal("Or did not close even though one input channel was already closed")
	}
}

// TestFirstToClose checks what return from Or func will be the fastest channel
func TestFirstToClose(t *testing.T) {
	start := time.Now()
	done := Or(
		makeDurationChan(10*time.Second),
		makeDurationChan(5*time.Second),
		makeDurationChan(50*time.Millisecond), // <-- fastest
		makeDurationChan(100*time.Millisecond),
	)

	select {
	case <-done:
		elapsed := time.Since(start)
		// Should complete in roughly 50 ms, not several seconds.
		if elapsed > 60*time.Millisecond {
			t.Fatalf("Or took too long: %v (expected around 50ms)", elapsed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Or timed out waiting for the fastest channel")
	}
}

func TestTwoChannels(t *testing.T) {
	done := Or(makeOpenChan(), makeClosedChan())
	select {
	case <-done:
		// Should be done instantly because we used closed chan
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Or(two channels) did not close in time")
	}
}

func TestLargeSet(t *testing.T) {
	const N = 16
	channels := make([]<-chan any, N)
	for i := range channels {
		channels[i] = makeOpenChan()
	}
	channels[N/2] = makeDurationChan(9 * time.Millisecond)

	done := Or(channels...)
	select {
	case <-done:
		// expected
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Timeout channel didn't finish in time, the Or func might not correctly takes channels")
	}
}

// -- Helpers --

func makeDurationChan(d time.Duration) <-chan any {
	c := make(chan any)
	go func() {
		defer close(c)
		time.Sleep(d)
	}()
	return c
}

func makeClosedChan() <-chan any {
	c := make(chan any)
	close(c)
	return c
}

func makeOpenChan() <-chan any {
	return make(chan any)
}
