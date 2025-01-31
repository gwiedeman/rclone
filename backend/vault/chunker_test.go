package vault

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestChunker(t *testing.T) {
	var cases = []struct {
		about          string
		data           string
		chunkSize      int64
		err            error
		expectedChunks []string
	}{
		{"invalid chunk size", "", 0, ErrInvalidChunkSize, []string{}},
		{"empty data", "", 1, nil, []string{}},
		{"empty data, larger chunk size", "", 2, nil, []string{}},
		{"smallest data", "a", 2, nil, []string{"a"}},
		{"data bigger than chunk, uneven", "abc", 2, nil, []string{"ab", "c"}},
		{"data bigger that chunk, even", "abcd", 2, nil, []string{"ab", "cd"}},
		{"smallest chunk size", "abcd", 1, nil, []string{"a", "b", "c", "d"}},
	}
	for _, c := range cases {
		f, err := os.CreateTemp(t.TempDir(), "vault-test-chunker*")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close() // nolint:errcheck
		if err = os.WriteFile(f.Name(), []byte(c.data), 0644); err != nil {
			t.Fatal(err)
		}
		ch, err := NewChunker(f.Name(), c.chunkSize)
		if err != nil && err == c.err {
			continue
		}
		if got := ch.NumChunks(); got != int64(len(c.expectedChunks)) {
			t.Fatalf("chunks: got %v, want %v", got, len(c.expectedChunks))
		}
		for i, ec := range c.expectedChunks {
			var buf bytes.Buffer
			chr := ch.ChunkReader(int64(i))
			if _, err = io.Copy(&buf, chr); err != nil {
				t.Fatalf("copy failed: %v", err)
			}
			if buf.String() != ec {
				t.Fatalf("got %v, want %v", buf.String(), ec)
			}
		}
	}
}

func TestChunkerChunkSize(t *testing.T) {
	var cases = []struct {
		about      string
		data       string
		chunkSize  int64
		err        error
		chunkSizes []int64
	}{
		{"uneven", "abcde", 2, nil, []int64{2, 2, 1}},
		{"even", "abcdef", 2, nil, []int64{2, 2, 2}},
		{"uneven, short", "abcdef", 4, nil, []int64{4, 2}},
	}
	for _, c := range cases {
		f, err := os.CreateTemp(t.TempDir(), "vault-test-chunker*")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close() // nolint:errcheck
		if err = os.WriteFile(f.Name(), []byte(c.data), 0644); err != nil {
			t.Fatal(err)
		}
		ch, err := NewChunker(f.Name(), c.chunkSize)
		if err != nil {
			t.Fatal(err)
		}
		var j int64
		for j = 0; j < ch.NumChunks(); j++ {
			got := ch.ChunkSize(j)
			if c.chunkSizes[j] != got {
				t.Fatalf("unexpected chunk size: got %v, want %v", got, c.chunkSizes[j])
			}
		}
	}
}
