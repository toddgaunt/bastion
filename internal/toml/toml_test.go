package toml_test

import (
	"bytes"
	"testing"

	"toddgaunt.com/monastery/internal/toml"
)

type childConfig struct {
	id         int
	name       string
	grandchild *childConfig
}

type parentConfig struct {
	id   int
	name string

	child childConfig
}

func TestMarshal(t *testing.T) {
	t.Parallel()

	example := parentConfig{
		id:   1,
		name: "mars",
		child: childConfig{
			id:   2,
			name: "romulus",
			grandchild: &childConfig{
				id:   3,
				name: "rome",
			},
		},
	}

	got, err := toml.Marshal(example)
	if err != nil {
		t.Fatalf("couldn't marshal structure: %v", err)
	}

	want := []byte("id = 1\nname = mars\n\n[child]\nid = 2\nname = romulus\n\n[child.grandchild]\nid = 3\nname = rome\n")

	if !bytes.Equal(got, want) {
		t.Fatalf("got bytes:\n%s\nwant bytes:\n%s\n", got, want)
	}
}
