package errors_test

import (
	"net/http"
	"testing"

	"github.com/toddgaunt/bastion/internal/errors"
)

const op = errors.Op("test-operation")

func TestE(t *testing.T) {
	testCases := []struct {
		name string
		err  errors.E

		wantStatus int
		wantKey    errors.Key
		wantErr    string
		wantMsg    string
	}{
		{
			name: "NoArgs",
			err:  errors.E{},

			wantErr: "",
			wantMsg: "",
		},
		{
			name: "OnlyOp",
			err: errors.E{
				Op: op,
			},

			wantErr: "test-operation",
			wantMsg: "",
		},
		{
			name: "WithOpAndStatus",
			err: errors.E{
				Op:   op,
				Code: http.StatusInternalServerError,
			},

			wantErr: "test-operation",
			wantMsg: "",
		},
		{
			name: "WithOpStatusAndKey",
			err: errors.E{
				Op:   op,
				Code: http.StatusInternalServerError,
				Key:  "TestInternalErr",
			},

			wantErr: "test-operation: TestInternalErr",
			wantMsg: "",
		},
		{
			name: "WithOpStatusKeyAndMessage",
			err: errors.E{
				Op:   op,
				Code: http.StatusUnauthorized,
				Key:  "TestUnauthorizedErr",
				Msg:  "account credentials aren't valid for any account",
			},

			wantErr: "test-operation: TestUnauthorizedErr",
			wantMsg: "account credentials aren't valid for any account",
		},
		{
			name: "WithAllFields",
			err: errors.E{
				Op:   op,
				Code: http.StatusBadRequest,
				Key:  "TestSerializationErr",
				Msg:  "request must be valid JSON",
				Err:  errors.New("failed to unmarshal json"),
			},

			wantErr: "test-operation: TestSerializationErr: failed to unmarshal json",
			wantMsg: "request must be valid JSON",
		},
		{
			name: "WithAllFieldsAndNestedError",
			err: errors.E{
				Op:   op,
				Code: http.StatusBadRequest,
				Key:  "TestSerializationErr",
				Msg:  "login failed",
				Err: &errors.E{
					Msg: "request must be valid JSON",
					Err: errors.New("failed to unmarshal json"),
				},
			},

			wantErr: "test-operation: TestSerializationErr: failed to unmarshal json",
			wantMsg: "login failed: request must be valid JSON",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.err

			if gotErr, wantErr := got.Error(), tc.wantErr; gotErr != wantErr {
				t.Errorf("errors don't match:\ngot:\n\t%v\nwant:\n\t%v", gotErr, wantErr)
			}

			if gotMsg, wantMsg := got.Message(), tc.wantMsg; gotMsg != wantMsg {
				t.Errorf("messages don't match:\ngot:\n\t%v\nwant:\n\t%v", gotMsg, wantMsg)
			}
		})
	}
}

func TestMsgf(t *testing.T) {
	got := errors.Msgf("Quick test %d", 1)
	if want := "Quick test 1"; string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
