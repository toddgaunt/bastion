package errors_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/toddgaunt/bastion/internal/errors"
)

func TestAnnotationWrap(t *testing.T) {
	err := errors.New("test error")

	testCases := []struct {
		name string
		ann  errors.Note

		want       string
		wantOp     string
		wantType   string
		wantTitle  string
		wantStatus int
		wantDetail string
		wantFields map[string]any
	}{
		{
			name: "operation",
			ann: errors.Note{
				Op: "operation",
			},

			want:   fmt.Sprintf("operation: %s", err.Error()),
			wantOp: "operation",
		},
		{
			name: "type",
			ann: errors.Note{
				Type: "type",
			},

			want:     fmt.Sprintf("type: %s", err.Error()),
			wantType: "type",
		},
		{
			name: "title",
			ann: errors.Note{
				Title: "title",
			},

			want:      fmt.Sprintf("title: %s", err.Error()),
			wantTitle: "title",
		},
		{
			name: "status",
			ann: errors.Note{
				StatusCode: 200,
			},

			want:       fmt.Sprintf("200: %s", err.Error()),
			wantStatus: 200,
		},
		{
			name: "detail",
			ann: errors.Note{
				Detail: "detail",
			},

			want:       err.Error(),
			wantDetail: "detail",
		},
		{
			name: "fields",
			ann: errors.Note{
				Fields: map[string]any{
					"string": "hello world",
					"bool":   true,
					"int":    25,
				},
			},

			want: err.Error(),
			wantFields: map[string]any{
				"string": "hello world",
				"bool":   true,
				"int":    25,
			},
		},
		{
			name: "allFields",
			ann: errors.Note{
				Op:         "operation",
				Type:       "type",
				Title:      "title",
				StatusCode: 400,
				Detail:     "detail",
				Fields: map[string]any{
					"string": "hello world",
					"bool":   true,
					"int":    25,
				},
			},

			want:       fmt.Sprintf("operation: type: title: 400: %s", err.Error()),
			wantOp:     "operation",
			wantType:   "type",
			wantTitle:  "title",
			wantStatus: 400,
			wantDetail: "detail",
			wantFields: map[string]any{
				"string": "hello world",
				"bool":   true,
				"int":    25,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.ann.Wrap(err)
			if got == nil {
				t.Fatalf("expected a non-nil error")
			}

			if got.Error() != tc.want {
				t.Fatalf("got error %s, want error %s", got, tc.want)
			}

			if got, want := got.Op(), tc.wantOp; got != want {
				t.Errorf("got op %s, want op %s", got, want)
			}

			if got, want := got.Type(), tc.wantType; got != want {
				t.Errorf("got type %s, want type %s", got, want)
			}

			if got, want := got.Title(), tc.wantTitle; got != want {
				t.Errorf("got title %s, want title %s", got, want)
			}

			if got, want := got.Status(), tc.wantStatus; got != want {
				t.Errorf("got status %d, want status %d", got, want)
			}

			if got, want := got.Detail(), tc.wantDetail; got != want {
				t.Errorf("got detail %s, want detail %s", got, want)
			}

			if got, want := got.Fields(), tc.wantFields; !reflect.DeepEqual(got, want) {
				t.Errorf("got fields %v, want fields %v", got, want)
			}
		})
	}
}

func TestNestedFields(t *testing.T) {
	want := map[string]any{
		"string": []any{"hello darkness", "my old field"},
		"bool":   []any{true, false},
		"int":    []any{24, 25},
	}

	err := errors.Note{
		Fields: map[string]any{
			"string": "hello darkness",
			"bool":   true,
			"int":    24,
		},
	}.Wrap(errors.Note{
		Fields: map[string]any{
			"string": "my old field",
			"bool":   false,
			"int":    25,
		},
	}.Wrap(errors.New("base")))

	got := err.Fields()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got fields %v, want fields %v", got, want)
	}
}
