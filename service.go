package tower

import "strings"

type Service struct {
	Name        string `json:"name,omitempty"`
	Environment string `json:"environment,omitempty"`
	Repository  string `json:"repository,omitempty"`
	Branch      string `json:"branch,omitempty"`
	Type        string `json:"type,omitempty"`
}

func (s Service) IsNil() bool {
	return s.Name == ""
}

type serviceWriteFlag uint8

const (
	noWrite     serviceWriteFlag = 0
	nameWritten serviceWriteFlag = 1 << iota
	typeWritten
)

func (s *serviceWriteFlag) Set(f serviceWriteFlag) {
	*s |= f
}

func (s serviceWriteFlag) Has(f serviceWriteFlag) bool {
	return s&f != 0
}

func (s Service) String() string {
	flag := noWrite
	builder := strings.Builder{}
	builder.Grow(len(s.Name) + len(s.Environment) + len(s.Type) + 2)
	if s.Name != "" {
		builder.WriteString(s.Name)
		flag.Set(nameWritten)
	}

	if s.Type != "" {
		if flag != noWrite {
			builder.WriteRune('-')
		}
		builder.WriteString(s.Type)
		flag.Set(typeWritten)
	}

	if s.Environment != "" {
		if flag != noWrite {
			builder.WriteRune('-')
		}
		builder.WriteString(s.Environment)
	}
	return builder.String()
}
