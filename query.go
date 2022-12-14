package tower

import "errors"

// Query is a namespace group that holds the tower's Query functions.
//
// Methods and functions under Query are utilities to search values in the error stack.
const Query query = 0

type query uint8

/*
GetHTTPCode Search for any error in the stack that implements HTTPCodeHint and return that value.

The API searches from the outermost error, and will return the first value it found.

Return 500 if there's no error that implements HTTPCodeHint in the stack.
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
GetCodeHint Search for any error in the stack that implements CodeHint and return that value.

The API searches from the outermost error, and will return the first value it found.

Return 500 if there's no error that implements CodeHint in the stack.

Used by Tower to search Code.
*/
func (query) GetCodeHint(err error) (code int) {
	if err == nil {
		return 500
	}

	if ch, ok := err.(CodeHint); ok { //nolint:errorlint
		return ch.Code()
	}

	return Query.GetCodeHint(errors.Unwrap(err))
}

/*
GetMessage Search for any error in the stack that implements MessageHint and return that value.

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
SearchCode Search the error stack for given code.

Given Code will be tested and the tower.Error is returned if:

 1. Any of the error in the stack implements CodeHint interface, matches the given code, and can be cast to tower.Error.
 2. Any of the error in the stack implements HTTPCodeHint interface, matches the given code, and can be cast to tower.Error.

Otherwise, this function will look deeper into the stack and
eventually returns nil when nothing in the stack implements those three and have the code.

The search operation is "Breath First", meaning the tower.Error is tested for CodeHint and HTTPCodeHint first before moving on.
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

	return Query.SearchCode(errors.Unwrap(err), code)
}

/*
SearchCodeHint Search the error stack for given code.

Given Code will be tested and the tower.Error is returned if any of the error in the stack implements CodeHint interface, matches the given code, and can be cast to tower.Error.

Otherwise, this function will look deeper into the stack and eventually returns nil when nothing in the stack implements CodeHint.
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
SearchHTTPCode Search the error stack for given code.

Given Code will be tested and the tower.Error is returned if any of the error in the stack implements HTTPCodeHint interface, matches the given code, and can be cast to tower.Error.

Otherwise, this function will look deeper into the stack and eventually returns nil when nothing in the stack implements HTTPCodeHint.
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

// CollectErrors Collects all the tower.Error in the error stack.
//
// It is sorted from the top most error to the bottom most error.
func (query) CollectErrors(err error) []Error {
	return collectErrors(err, nil)
}

func collectErrors(err error, input []Error) []Error {
	if err == nil {
		return input
	}

	if err, ok := err.(Error); ok { //nolint:errorlint
		input = append(input, err)
	}

	return collectErrors(errors.Unwrap(err), input)
}

// GetStack Gets the error stack by checking CallerHint.
//
// Tower recursively checks the given error if it implements CallerHint until all the error in the stack are checked.
//
// If you wish to get list of tower.Error use CollectErrors instead.
func (query) GetStack(err error) []KeyValue[Caller, error] {
	in := make([]KeyValue[Caller, error], 0, 10)
	return getStackList(err, in)
}

func getStackList(err error, input []KeyValue[Caller, error]) []KeyValue[Caller, error] {
	if err == nil {
		return input
	}
	if ch, ok := err.(CallerHint); ok { //nolint:errorlint
		return append(input, NewKeyValue(ch.Caller(), err))
	}
	return getStackList(errors.Unwrap(err), input)
}

// TopError Gets the outermost tower.Error instance in the error stack.
// Returns nil if no tower.Error instance found in the stack.
func (query) TopError(err error) Error {
	if err == nil {
		return nil
	}

	if e, ok := err.(Error); ok { //nolint:errorlint
		return e
	}

	return Query.TopError(errors.Unwrap(err))
}

// BottomError Gets the outermost tower.Error instance in the error stack.
// Returns nil if no tower.Error instance found in the stack.
func (query) BottomError(err error) Error {
	top := Query.TopError(err)
	if top == nil {
		return nil
	}
	var result Error
	unwrapped := top.Unwrap()
	for unwrapped != nil {
		if e, ok := unwrapped.(Error); ok { //nolint:errorlint
			result = e
		}
		unwrapped = errors.Unwrap(err)
	}

	return result
}

// Cause returns the root cause.
func (query) Cause(err error) error {
	unwrapped := errors.Unwrap(err)
	for unwrapped != nil {
		err = unwrapped
		unwrapped = errors.Unwrap(err)
	}
	return err
}
