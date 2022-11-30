package tower

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
)

type TestingJSONLogger struct {
	buf *bytes.Buffer
	mu  sync.Mutex
}

// NewTestingJSONLogger returns a very basic TestingJSONLogger.
func NewTestingJSONLogger() *TestingJSONLogger {
	return &TestingJSONLogger{
		buf: new(bytes.Buffer),
	}
}

// Log implements tower.Logger.
func (t *TestingJSONLogger) Log(ctx context.Context, entry Entry) {
	t.mu.Lock()
	defer t.mu.Unlock()

	err := json.NewEncoder(t.buf).Encode(entry)
	if err != nil {
		t.buf.WriteString(err.Error())
	}
}

// LogError implements tower.Logger.
func (t *TestingJSONLogger) LogError(ctx context.Context, err Error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	errJson := json.NewEncoder(t.buf).Encode(err)
	if errJson != nil {
		t.buf.WriteString(errJson.Error())
	}
}

// Reset resets the buffer to be empty.
func (t *TestingJSONLogger) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf.Reset()
}

// String returns the accumulated bytes as string.
func (t *TestingJSONLogger) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.buf.String()
}

// Bytes returns the accumulated bytes.
func (t *TestingJSONLogger) Bytes() []byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.buf.Bytes()
}
