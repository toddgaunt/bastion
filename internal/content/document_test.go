package content_test

import (
	"errors"
	"testing"

	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/gmath"
	"github.com/toddgaunt/bastion/internal/tests"
)

func TestMarshalDocument(t *testing.T) {
	testCases := []struct {
		name     string
		document content.Document

		want string
		err  error
	}{
		{
			name: "test name",
			document: content.Document{
				Properties: content.Properties{
					"Title":  {"Lord of the Rings", "The Fellowship of the Ring"},
					"Author": {"Tolkien"},
					"Tag":    {"Adventure", "Fantasy", "Myth"},
				},
				Format:  "markdown",
				Content: []byte("The weary ring bearer traveled on,\nnot knowing where he was to go."),
			},

			want: gmath.Concat(
				"Author: Tolkien\n",
				"Tag: Adventure\n",
				"Tag: Fantasy\n",
				"Tag: Myth\n",
				"Title: Lord of the Rings\n",
				"Title: The Fellowship of the Ring\n",
				"=== markdown ===\n",
				"The weary ring bearer traveled on,\nnot knowing where he was to go.",
			),
			err: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := content.MarshalDocument(tc.document)
			if !errors.Is(err, tc.err) {
				t.Fatalf("got error %v, want error %v", err, tc.err)
			}

			if got := string(got); got != tc.want {
				//t.Fatalf("Document doesn't match what was expected:\ngot:\n%q\nwant:\n%q", got, tc.want)

				t.Fatalf("Document doesn't match what was expected:\n%v",
					string(tests.Diff(tc.want, got)),
				)
			}
		})
	}
}

func TestMarshalUnmarshalDocument(t *testing.T) {
}
