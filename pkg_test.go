package hensu_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/my-rv/hensu"
)

func TestFacade_SetRead(t *testing.T) {
	fs := hensu.NewMemFileSystem()
	s := hensu.NewStore("/facade.env", hensu.WithFileSystem(fs))
	if err := s.Set(context.Background(), map[string]string{"FOO": "bar"}); err != nil {
		t.Fatal(err)
	}
	v, err := s.Read(context.Background(), "FOO")
	if err != nil || v != "bar" {
		t.Fatalf("%q %v", v, err)
	}
}

func ExampleStore() {
	fs := hensu.NewMemFileSystem()
	s := hensu.NewStore("/app/.env", hensu.WithFileSystem(fs))
	_ = s.Set(context.Background(), map[string]string{"INTERNAL:FOO": "bar"})
	v, err := s.Read(context.Background(), "INTERNAL__FOO")
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
	// Output: bar
}

func ExampleCanonicalKey() {
	fmt.Println(hensu.CanonicalKey("internal:sessions:foo"))
	// Output: INTERNAL__SESSIONS__FOO
}
