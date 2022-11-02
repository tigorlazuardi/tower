package tower_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/tigorlazuardi/tower"
)

type errorTest struct {
	Foo string
}

func (e errorTest) Error() string {
	return "errorTest.Foo"
}

type errorTest2 struct{}

func (e errorTest2) Error() string {
	return strings.Repeat("boo", 20)
}

type jsonTest struct{}

func (j jsonTest) MarshalJSON() ([]byte, error) {
	return []byte(`{"foo":"bar",  "baz": "baz"}`), nil
}

type displayTest string

// Display returns a human readable and rich with information for the implementer.
func (d displayTest) Display() string {
	return strings.Repeat("foo", 3)
}

func TestFields_WriteDisplay(t *testing.T) {
	type args struct {
		w tower.LineWriter
	}
	tests := []struct {
		name string
		f    tower.Fields
		args args
		want string
	}{
		{
			name: "expected result - simple",
			f: map[string]any{
				"foo": "bar",
				"wtf": 123,
			},
			args: args{
				w: tower.NewLineWriterBuilder().Separator("\n").Build(),
			},
			want: "foo: bar\nwtf: 123",
		},
		{
			name: "expected result - complex",
			f: map[string]any{
				"display_writer": tower.Fields{
					"bar":   2000,
					"baz":   "www",
					"bytes": bytes.Repeat([]byte(`baz`), 20),
					"buzz": tower.Fields{
						"light": "year",
					},
				},
				"json":        []byte(`{"foo":"bar"}`),
				"bytes":       bytes.Repeat([]byte(`baz`), 20),
				"display":     displayTest(""),
				"struct":      jsonTest{},
				"errors.new":  errors.New("foo"),
				"error_test":  errorTest{},
				"error_test2": errorTest2{},
				"function":    func() {},
			},
			args: args{
				w: tower.NewLineWriterBuilder().Separator("\n").Prefix(">> ").Build(),
			},
			want: strings.TrimSpace(`
>> bytes:
>>     bazbazbazbazbazbazbazbazbazbazbazbazbazbazbazbazbazbazbazbaz
>> display: foofoofoo
>> display_writer:
>>     bar: 2000
>>     baz: www
>>     buzz:
>>         light: year
>>     bytes:
>>         bazbazbazbazbazbazbazbazbazbazbazbazbazbazbazbazbazbazbazbaz
>> error_test:
>>     {
>>         "Foo": ""
>>     }
>> error_test_summary: errorTest.Foo
>> error_test2:
>>     booboobooboobooboobooboobooboobooboobooboobooboobooboobooboo
>> errors.new: foo
>> function:
>>     json: unsupported type: func()
>> json:
>>     {
>>         "foo": "bar"
>>     }
>> struct:
>>     {
>>         "foo": "bar",
>>         "baz": "baz"
>>     }
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.WriteDisplay(tt.args.w)
			got := tt.args.w.String()
			if got != tt.want {
				t.Errorf("got and want is not equal.\n\nwant:\n%s\n\ngot:\n%s\n", tt.want, got)
			}
		})
	}
}
