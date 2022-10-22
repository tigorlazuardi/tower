package tower

import "errors"

/*
Search for any error in the stack that implements HTTPCodeHint and return that value.

The API searches from the outermost error, and will return the first value it found.

Return 500 if there's no error that implements HTTPCodeHint in the stack.

Used by Tower to search HTTP Code.
*/
func GetHTTPCode(err error) (code int) {
	if err == nil {
		return 500
	}

	if ch, ok := err.(HTTPCodeHint); ok { //nolint:errorlint
		return ch.HTTPCode()
	}

	return GetHTTPCode(errors.Unwrap(err))
}

/*
Search for any error in the stack that implements BodyCodeHint and return that value.

The API searches from the outermost error, and will return the first value it found.

Return 5500 if there's no error that implements BodyCodeHint in the stack.

Used by Tower to search Body Code.
*/
func GetBodyCode(err error) (code int) {
	if err == nil {
		return 5500
	}

	if ch, ok := err.(BodyCodeHint); ok { //nolint:errorlint
		return ch.BodyCode()
	}

	return GetBodyCode(errors.Unwrap(err))
}

/*
Search for any error in the stack that implements CodeHint and return that value.

The API searches from the outermost error, and will return the first value it found.

Return 5500 if there's no error that implements CodeHint in the stack.

Used by Tower to search Body Code.
*/
func GetCodeHint(err error) (code int) {
	if err == nil {
		return 5500
	}

	if ch, ok := err.(CodeHint); ok { //nolint:errorlint
		return ch.Code()
	}

	return GetCodeHint(errors.Unwrap(err))
}

/*
Search for any error in the stack that implements MessageHint and return that value.

The API searches from the outermost error, and will return the first value it found.

Return empty string if there's no error that implements MessageHint in the stack.

Used by Tower to search Message in the error.
*/
func GetMessage(err error) (message string) {
	if err == nil {
		return ""
	}

	if ch, ok := err.(MessageHint); ok { //nolint:errorlint
		return ch.Message()
	}

	return GetMessage(errors.Unwrap(err))
}

/*
Search the error stack for given code.

Given Code will be tested and the tower.Error is returned if:

 1. Any of the error in the stack implements CodeHint interface, matches the given code, and can be casted to tower.Error.
 2. Any of the error in the stack implements HTTPCodeHint interface, matches the given code, and can be casted to tower.Error.
 3. Any of the error in the stack implements BodyCodeHint interface, matches the given code, and can be casted to tower.Error.

Otherwise this function will look deeper into the stack and eventually returns nil when nothing in the stack implements those three and have the code.
*/
func SearchCode(err error, code int) Error {
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

	return SearchCode(errors.Unwrap(err), code)
}

/*
Search the error stack for given code.

Given Code will be tested and the tower.Error is returned if any of the error in the stack implements CodeHint interface, matches the given code, and can be casted to tower.Error.

Otherwise this function will look deeper into the stack and eventually returns nil when nothing in the stack implements CodeHint.
*/
func SearchCodeHint(err error, code int) Error {
	if err == nil {
		return nil
	}

	if ch, ok := err.(CodeHint); ok && ch.Code() == code { //nolint:errorlint
		if err, ok := err.(Error); ok { //nolint:errorlint
			return err
		}
	}

	return SearchCodeHint(errors.Unwrap(err), code)
}

/*
Search the error stack for given code.

Given Code will be tested and the tower.Error is returned if any of the error in the stack implements HTTPCodeHint interface, matches the given code, and can be casted to tower.Error.

Otherwise this function will look deeper into the stack and eventually returns nil when nothing in the stack implements HTTPCodeHint.
*/
func SearchHTTPCode(err error, code int) Error {
	if err == nil {
		return nil
	}

	if ch, ok := err.(HTTPCodeHint); ok && ch.HTTPCode() == code { //nolint:errorlint
		if err, ok := err.(Error); ok { //nolint:errorlint
			return err
		}
	}

	return SearchHTTPCode(errors.Unwrap(err), code)
}

/*
Search the error stack for given code.

Given Code will be tested and the tower.Error is returned if any of the error in the stack implements BodyCodeHint interface, matches the given code, and can be casted to tower.Error.

Otherwise this function will look deeper into the stack and eventually returns nil when nothing in the stack implements BodyCodeHint.
*/
func SearchBodyCode(err error, code int) Error {
	if err == nil {
		return nil
	}

	if ch, ok := err.(BodyCodeHint); ok && ch.BodyCode() == code { //nolint:errorlint
		if err, ok := err.(Error); ok { //nolint:errorlint
			return err
		}
	}

	return SearchHTTPCode(errors.Unwrap(err), code)
}
