package towerdiscord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/kinbiko/jsonassert"
	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/bucket"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockBucket map[string]any

func (m mockBucket) Upload(ctx context.Context, files []bucket.File) []bucket.UploadResult {
	results := make([]bucket.UploadResult, len(files))
	for i, f := range files {
		url := "https://example.com/" + f.Filename()
		results[i] = bucket.UploadResult{
			File: f,
			URL:  "https://example.com/" + f.Filename(),
		}
		m[url] = results[i]
	}
	return results
}

func TestBucket(t *testing.T) {
	tests := []struct {
		name       string
		test       func(t *testing.T) callback
		testBucket func(t *testing.T, b mockBucket)
		wantCount  int
		error      error
		message    string
		context    []any
		extraOpts  []DiscordOption
	}{
		{
			name: "should upload file",
			test: func(t *testing.T) callback {
				return func(r *http.Request) {
					if r.Method != http.MethodPost {
						t.Errorf("want method %s, got %s", http.MethodPost, r.Method)
					}
					if r.Header.Get("Content-Type") != "application/json" {
						t.Errorf("want content type %s, got %s", "application/json", r.Header.Get("Content-Type"))
					}
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatalf("failed to read body: %v", err)
					}
					j := jsonassert.New(t)
					j.Assertf(string(body), `
					{
					  "content": "@here an error has occurred on service **test** on type **test** on environment **test**",
					  "embeds": [
						{
						  "title": "Summary",
						  "type": "rich",
						  "description": "<<PRESENCE>>",
						  "color": 1606980
						},
						{
						  "title": "Error",
						  "type": "rich",
						  "description": "<<PRESENCE>>",
						  "color": 7405835
						},
						{
							"title": "Error Stack",
							"type": "rich",
							"description": "<<PRESENCE>>",
							"color": 6098454
						},
						{
							"title": "Metadata",
							"type": "rich",
							"description": "<<PRESENCE>>",
							"timestamp": "<<PRESENCE>>",
							"color": 6576731,
							"fields": [
									{
										"name": "Service",
										"value": "test",
										"inline": true
									},
									{
										"name": "Type",
										"value": "test",
										"inline": true
									},
									{
										"name": "Environment",
										"value": "test",
										"inline": true
									},
									{
										"name": "Thread ID",
										"value": "<<PRESENCE>>",
										"inline": true
									},
									{
										"name": "Message Iteration",
										"value": "1",
										"inline": true
									},
									{
										"name": "Next Possible Earliest Repeat",
										"value": "<<PRESENCE>>"
									}
								]
							}
						],
						"attachments": [
							{
								"id": 0,
								"filename": "<<PRESENCE>>",
								"content_type": "text/markdown; charset=utf-8",
								"size": 9029,
								"url": "<<PRESENCE>>"
							},
							{
								"id": 1,
								"filename": "<<PRESENCE>>",
								"content_type": "application/json",
								"size": "<<PRESENCE>>",
								"url": "<<PRESENCE>>"
							}
						]
					}`)
					if t.Failed() {
						out := new(bytes.Buffer)
						_ = json.Indent(out, body, "", "  ")
						fmt.Println(out.String())
					}
				}
			},
			testBucket: func(t *testing.T, b mockBucket) {
				if len(b) != 2 {
					t.Errorf("want 2, got %d", len(b))
					if t.Failed() {
						out := new(bytes.Buffer)
						enc := json.NewEncoder(out)
						enc.SetEscapeHTML(true)
						_ = enc.Encode(b)
						fmt.Printf(out.String())
					}
				}
			},
			wantCount: 1,
			error:     errors.New(strings.Repeat("foo", 3000)),
			message:   "",
			context:   nil,
			extraOpts: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mBucket := mockBucket{}
			tow := tower.NewTower(tower.Service{Name: "test", Environment: "test", Type: "test"})
			m := newMockClient(t, tt.test(t))
			defer m.Close()
			d := NewDiscordBot("", append([]DiscordOption{WithClient(m), WithBucket(mBucket)}, tt.extraOpts...)...)
			tow.RegisterMessenger(d)
			if tt.error != nil {
				_ = tow.Wrap(tt.error).Message(tt.message).Context(tt.context...).Notify(context.Background())
			} else {
				_ = tow.NewEntry(tt.message).Context(tt.context...).Notify(context.Background())
			}
			err := tow.Wait(context.Background())
			if err != nil {
				t.Fatalf("tower.Wait() error = %v", err)
			}
			m.Wait()
			if m.count != tt.wantCount {
				t.Errorf("m.count = %d, want = %d", m.count, tt.wantCount)
			}
			tt.testBucket(t, mBucket)
		})
	}
}
