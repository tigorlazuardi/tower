package towerhttp

import (
	"github.com/tigorlazuardi/tower"
	"io"
	"net/http"
	"testing"
)

func TestResponder_RespondStream(t *testing.T) {
	type fields struct {
		encoder          Encoder
		transformer      BodyTransformer
		errorTransformer ErrorBodyTransformer
		tower            *tower.Tower
		compressor       Compressor
		callerDepth      int
		hooks            []RespondHook
	}
	type args struct {
		rw          http.ResponseWriter
		request     *http.Request
		contentType string
		body        io.Reader
		opts        []RespondOption
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Responder{
				encoder:          tt.fields.encoder,
				transformer:      tt.fields.transformer,
				errorTransformer: tt.fields.errorTransformer,
				tower:            tt.fields.tower,
				compressor:       tt.fields.compressor,
				callerDepth:      tt.fields.callerDepth,
				hooks:            tt.fields.hooks,
			}
			r.RespondStream(tt.args.rw, tt.args.request, tt.args.contentType, tt.args.body, tt.args.opts...)
		})
	}
}
