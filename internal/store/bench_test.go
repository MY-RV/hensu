package store_test

import (
	"context"
	"testing"

	"github.com/my-rv/hensu/internal/store"
)

func BenchmarkStore_SetGet(b *testing.B) {
	fs := store.NewMemFileSystem()
	s := store.NewStore("/bench.env", store.WithFileSystem(fs))
	ctx := context.Background()
	_ = s.Set(ctx, map[string]string{"SEED": "1"})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := s.Set(ctx, map[string]string{"K": "v"}); err != nil {
			b.Fatal(err)
		}
		if _, err := s.Read(ctx, "K"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCanonicalKey(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = store.CanonicalKey("internal:sessions:max_minutes_timeout")
	}
}
