package tower

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
