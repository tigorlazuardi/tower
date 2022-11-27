package tower_test

import (
	"errors"
	"fmt"
	"github.com/tigorlazuardi/tower"
	"strings"
	"testing"
)

func Test_Error_WriteError(t *testing.T) {
	tests := []struct {
		name   string
		error  tower.Error
		writer func() (tower.LineWriter, fmt.Stringer)
		want   string
	}{
		{
			name: "No Duplicates",
			error: func() tower.Error {
				err := tower.BailFreeze("bail")
				err = tower.WrapFreeze(err, "wrap")
				return tower.Wrap(err).Freeze()
			}(),
			writer: func() (tower.LineWriter, fmt.Stringer) {
				s := &strings.Builder{}
				lw := tower.NewLineWriter(s).LineBreak(" => ").Build()
				return lw, s
			},
			want: "wrap => bail",
		},
		{
			name: "No Duplicates - Tail",
			error: func() tower.Error {
				err := errors.New("errors.New")
				err = tower.WrapFreeze(err, "wrap")
				err = tower.Wrap(err).Freeze()
				return tower.Wrap(err).Message("foo").Freeze()
			}(),
			writer: func() (tower.LineWriter, fmt.Stringer) {
				s := &strings.Builder{}
				lw := tower.NewLineWriter(s).LineBreak(" => ").Build()
				return lw, s
			},
			want: "foo => wrap => errors.New",
		},
		{
			name: "Ensure different messages are written",
			error: func() tower.Error {
				err := tower.BailFreeze("bail")
				err = tower.WrapFreeze(err, "wrap")
				return tower.Wrap(err).Message("wrap 2").Freeze()
			}(),
			writer: func() (tower.LineWriter, fmt.Stringer) {
				s := &strings.Builder{}
				lw := tower.NewLineWriter(s).LineBreak(" => ").Build()
				return lw, s
			},
			want: "wrap 2 => wrap => bail",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, buf := tt.writer()
			tt.error.WriteError(writer)
			if got := buf.String(); got != tt.want {
				t.Errorf("Error.WriteError() = %v, want %v", got, tt.want)
			}
		})
	}
}
