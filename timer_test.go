package yamux

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestDeadlineTimerReuse(t *testing.T) {
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()

	cs, err := client.OpenStream()
	if err != nil {
		t.Fatal(err)
	}
	ss, err := server.AcceptStream()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		_ = ss.SetReadDeadline(time.Now().Add(20 * time.Millisecond))
		if _, err := ss.Read(make([]byte, 1)); !errors.Is(err, ErrTimeout) && !errors.Is(err, os.ErrDeadlineExceeded) {
			t.Fatalf("round %d: want timeout, got %v", i, err)
		}
		_ = ss.SetReadDeadline(time.Now().Add(5 * time.Second))
		if _, err := cs.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
		if n, err := ss.Read(make([]byte, 1)); n != 1 || err != nil {
			t.Fatalf("round %d: read %d %v", i, n, err)
		}
	}
}

func BenchmarkDeadlineStream(b *testing.B) {
	client, server := testClientServer(b)
	cs, _ := client.OpenStream()
	ss, _ := server.AcceptStream()
	go func() {
		buf := make([]byte, 64)
		for {
			if _, err := ss.Read(buf); err != nil {
				return
			}
		}
	}()
	_ = cs.SetWriteDeadline(time.Now().Add(time.Hour))
	p := make([]byte, 64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := cs.Write(p); err != nil {
			b.Fatal(err)
		}
	}
}
