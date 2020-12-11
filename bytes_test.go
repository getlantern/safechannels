package safechannels

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBytes(t *testing.T) {
	ch := NewBytes(1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		require.True(t, ch.Write([]byte("hello")))
		require.True(t, ch.Write([]byte("world")))
		ch.Write([]byte("good")) // may write or not, depending on timing
		require.False(t, ch.Write([]byte("bye")))
		// Wait for channel to have been closed
		time.Sleep(250 * time.Millisecond)
		// Drain unread writes
		for range ch.Read() {
		}
		// Write again, should fail immediately even though there's room on channel
		require.False(t, ch.Write([]byte("really")))
	}()

	b := <-ch.Read()
	require.Equal(t, "hello", string(b))
	b = <-ch.Read()
	require.Equal(t, "world", string(b))
	ch.Close()
	b, ok := <-ch.Read()
	if ok {
		require.Equal(t, "good", string(b))
		b, ok = <-ch.Read()
		require.False(t, ok)
	} else {
		// due to timing, "good" was never written
	}

	wg.Wait()
}

func BenchmarkSafeWrites(b *testing.B) {
	val := []byte("value")
	ch := NewBytes(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.Write(val)
	}
}
func BenchmarkDirectWrites(b *testing.B) {
	val := []byte("value")
	ch := make(chan []byte, b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- val
	}
}
