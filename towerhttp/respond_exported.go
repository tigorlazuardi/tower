package towerhttp

import (
	"context"
	"io"
	"net/http"
)

var exportedResponder *Responder

func init() {
	exportedResponder = NewResponder()
	exportedResponder.SetCallerDepth(4)
}

// Responder returns the global responder instance.
func (exported) Responder() *Responder {
	return exportedResponder
}

// SetResponder sets the global responder instance.
func (exported) SetResponder(r *Responder) {
	exportedResponder = r
}

// Respond with the given body and options.
//
// body is expected to be a serializable type. For streams, use RespondStream.
//
// HTTP status by default is http.StatusOK. If body implements tower.HTTPCodeHint, the status code will be set to the
// value returned by the tower.HTTPCodeHint method. If the towerhttp.RO.StatusCode RespondOption is set, it will override
// the status regardless of the tower.HTTPCodeHint.
//
// There's a special case if you pass http.NoBody as body, there will be no respond body related operations executed.
// StatusCode default value is STILL http.StatusOK. If you wish to set the status code to http.StatusNoContent, you
// can still override the output by setting the related RespondOption.
//
// Body of nil has different treatment with http.NoBody. if body is nil, the nil value is still passed to the BodyTransformer implementer,
// therefore the final result body may not actually be empty.
func Respond(ctx context.Context, rw http.ResponseWriter, body any, opts ...RespondOption) {
	exportedResponder.Respond(ctx, rw, body, opts...)
}

// RespondStream writes the given stream to the http.ResponseWriter.
//
// If the stream implements tower.HTTPCodeHint, the status code will be set to the value returned by the tower.HTTPCodeHint.
//
// If the Compression supports StreamCompression, the stream will be compressed by said StreamCompression and
// written to the http.ResponseWriter.
//
// There's a special case if you pass http.NoBody as body, there will be no respond body related operations executed.
// StatusCode default value is STILL http.StatusOK. You can still override this output by setting the
// related RespondOption.
// With http.NoBody as body, Towerhttp will immediately respond with status code after RespondOption are evaluated
// and end the process.
//
// Body of nil will be treated as http.NoBody.
func RespondStream(ctx context.Context, rw http.ResponseWriter, contentType string, body io.Reader, opts ...RespondOption) {
	exportedResponder.RespondStream(ctx, rw, contentType, body, opts...)
}

// RespondError writes the given error to the http.ResponseWriter.
//
// error is expected to be a serializable type.
//
// HTTP Status code by default is http.StatusInternalServerError. If error implements tower.HTTPCodeHint, the status code will be set to the
// value returned by the tower.HTTPCodeHint method. If the towerhttp.RO.StatusCode RespondOption is set, it will override
// the status regardless of the tower.HTTPCodeHint.
//
// if err is nil, it will be replaced with "Internal Server Error" message. It is done this way, because the library
// assumes that you mishandled the method and to prevent sending empty values, a generic Internal Server Error message
// will be sent instead. If you wish to send an empty response, use Respond with http.NoBody as body.
func RespondError(ctx context.Context, rw http.ResponseWriter, err error, opts ...RespondOption) {
	exportedResponder.RespondError(ctx, rw, err, opts...)
}
