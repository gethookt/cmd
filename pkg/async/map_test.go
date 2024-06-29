package async_test

import (
	"sync"
	"testing"

	"hookt.dev/cmd/pkg/async"

	"golang.org/x/sync/errgroup"
)

func slice(len, val int) []int {
	s := make([]int, len+1)
	s[0] = val
	return s
}

func TestMap(t *testing.T) {
	var (
		m  async.Map
		eg errgroup.Group
		wg sync.WaitGroup
	)

	cases := map[string][]int{
		"key1": slice(16, 1),
		"key2": slice(32, 2),
		"key3": slice(8, 3),
		"key4": slice(48, 4),
		"key5": slice(64, 5),
		"key6": slice(4, 6),
		"key7": slice(128, 7),
		"key8": slice(512, 8),
		"key9": slice(256, 9),
	}

	for key, want := range cases {
		for i := range want[1:] {
			wg.Add(1)
			eg.Go(func() error {
				wg.Done()
				v, _ := m.Load(key)
				if v, ok := v.(int); ok {
					want[i+1] = v
				}
				return nil
			})

		}
	}

	wg.Wait()

	for key, want := range cases {
		m.Store(key, want[0])
	}

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}

	for _, want := range cases {
		for i, v := range want[1:] {
			if v != want[0] {
				t.Errorf("want %v, got %v at index %v", want[0], v, i)
			}
		}
	}
}
