package tower

import "errors"

// Groupings for Query and Query functions.
//
// Methods and functions under Query are utilities to search values in the error stack.
const Query query = 0

type query uint8

/*
Search for any error in the stack that implements HTTPCodeHint and return that value.

The API searches from the outermost error, and will return the first value it found.

Return 500 if there's no error that implements HTTPCodeHint in the stack.

Used by Tower to search HTTP Code.
*/
func (query) GetHTTPCode(err error) (code int) {
	if err == nil {
		return 500
	}

	if ch, ok := err.(HTTPCodeHint); ok { //nolint:errorlint
		return ch.HTTPCode()
	}

	return Query.GetHTTPCode(errors.Unwrap(err))
}

/*
Search for any error in the stack that implements BodyCodeHint and return that value.

The API searches from the outermost error, and will return the first value it found.

Return 5500 if there's no error that implements BodyCodeHint in the stack.

Used by Tower to search Body Code.
*/
func (query) GetBodyCode(err error) (code int) {
	if err == nil {
		return 5500
	}

	if ch, ok := err.(BodyCodeHint); ok { //nolint:errorlint
		return ch.BodyCode()
	}

	return Query.GetBodyCode(errors.Unwrap(err))
}

/*
Search for any error in the stack that implements CodeHint and return that value.

The API searches from the outermost error, and will return the first value it found.

Return 5500 if there's no error that implements CodeHint in the stack.

Used by Tower to search Body Code.
*/
func (query) GetCodeHint(err error) (code int) {
	if err == nil {
		return 5500
	}

	if ch, ok := err.(CodeHint); ok { //nolint:errorlint
		return ch.Code()
	}

	return Query.GetCodeHint(errors.Unwrap(err))
}

/*
Search for any error in the stack that implements MessageHint and return that value.

The API searches from the outermost error, and will return the first value it found.

Return empty string if there's no error that implements MessageHint in the stack.

Used by Tower to search Message in the error.
*/
func (query) GetMessage(err error) (message string) {
	if err == nil {
		return ""
	}

	if ch, ok := err.(MessageHint); ok { //nolint:errorlint
		return ch.Message()
	}

	return Query.GetMessage(errors.Unwrap(err))
}

/*
Search the error stack for given code.

Given Code will be tested and the tower.Error is returned if:

 1. Any of the error in the stack implements CodeHint interface, matches the given code, and can be casted to tower.Error.
 2. Any of the error in the stack implements HTTPCodeHint interface, matches the given code, and can be casted to tower.Error.
 3. Any of the error in the stack implements BodyCodeHint interface, matches the given code, and can be casted to tower.Error.

Otherwise this function will look deeper into the stack and eventually returns nil when nothing in the stack implements those three and have the code.
*/
func (query) SearchCode(err error, code int) Error {
	if err == nil {
		return nil
	}

	// It's important to not use the other Search API for brevity.
	// This is because we are

	if ch, ok := err.(CodeHint); ok && ch.Code() == code { //nolint:errorlint
		if err, ok := err.(Error); ok { //nolint:errorlint
			return err
		}
	}

	if ch, ok := err.(HTTPCodeHint); ok && ch.HTTPCode() == code { //nolint:errorlint
		if err, ok := err.(Error); ok { //nolint:errorlint
			return err
		}
	}

	if ch, ok := err.(BodyCodeHint); ok && ch.BodyCode() == code { //nolint:errorlint
		if err, ok := err.(Error); ok { //nolint:errorlint
			return err
		}
	}

	return Query.SearchCode(errors.Unwrap(err), code)
}

/*
Search the error stack for given code.

Given Code will be tested and the tower.Error is returned if any of the error in the stack implements CodeHint interface, matches the given code, and can be casted to tower.Error.

Otherwise this function will look deeper into the stack and eventually returns nil when nothing in the stack implements CodeHint.
*/
func (query) SearchCodeHint(err error, code int) Error {
	if err == nil {
		return nil
	}

	if ch, ok := err.(CodeHint); ok && ch.Code() == code { //nolint:errorlint
		if err, ok := err.(Error); ok { //nolint:errorlint
			return err
		}
	}

	return Query.SearchCodeHint(errors.Unwrap(err), code)
}

/*
Search the error stack for given code.

Given Code will be tested and the tower.Error is returned if any of the error in the stack implements HTTPCodeHint interface, matches the given code, and can be casted to tower.Error.

Otherwise this function will look deeper into the stack and eventually returns nil when nothing in the stack implements HTTPCodeHint.
*/
func (query) SearchHTTPCode(err error, code int) Error {
	if err == nil {
		return nil
	}

	if ch, ok := err.(HTTPCodeHint); ok && ch.HTTPCode() == code { //nolint:errorlint
		if err, ok := err.(Error); ok { //nolint:errorlint
			return err
		}
	}

	return Query.SearchHTTPCode(errors.Unwrap(err), code)
}

/*
Search the error stack for given code.

Given Code will be tested and the tower.Error is returned if any of the error in the stack implements BodyCodeHint interface, matches the given code, and can be casted to tower.Error.

Otherwise this function will look deeper into the stack and eventually returns nil when nothing in the stack implements BodyCodeHint.
*/
func (query) SearchBodyCode(err error, code int) Error {
	if err == nil {
		return nil
	}

	if ch, ok := err.(BodyCodeHint); ok && ch.BodyCode() == code { //nolint:errorlint
		if err, ok := err.(Error); ok { //nolint:errorlint
			return err
		}
	}

	return Query.SearchBodyCode(errors.Unwrap(err), code)
}

// Gets the error stack by checking CallerHint.
//
// Tower recursively checks the given error if it implements CallerHint until all the error in the stack are checked.
func (query) GetStack(err error) []Caller {
	in := make([]Caller, 0, 10)
	return getStackList(err, in)
}

func getStackList(err error, input []Caller) []Caller {
	if err == nil {
		return input
	}
	if ch, ok := err.(CallerHint); ok { //nolint:errorlint
		return append(input, ch.Caller())
	}
	return getStackList(errors.Unwrap(err), input)
}

// Gets the outermost tower.Error instance in the error stack.
// Returns nil if no tower.Error instance found in the stack.
func (query) GetError(err error) Error {
	if err == nil {
		return nil
	}

	if e, ok := err.(Error); ok { //nolint:errorlint
		return e
	}

	return Query.GetError(errors.Unwrap(err))
}
