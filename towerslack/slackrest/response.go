package slackrest

type ErrorResponse struct {
	Ok  bool   `json:"ok"`
	Err string `json:"error"`
}

func (err *ErrorResponse) Error() string {
	return err.Err
}
