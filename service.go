package tower

import "strings"

type Service struct {
	Name        string
	Environment string
	Repository  string
	Branch      string
	Type        string
}

func (s Service) IsNil() bool {
	return s.Name == ""
}

func (s Service) String() string {
	builder := strings.Builder{}
	builder.Grow(len(s.Name) + len(s.Environment) + len(s.Type) + 2)
	builder.WriteString(s.Name)
	builder.WriteRune('-')
	builder.WriteString(s.Type)
	builder.WriteRune('-')
	builder.WriteString(s.Environment)
	return builder.String()
}
