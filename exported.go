package tower

// Wrap the error. The returned ErrorBuilder may be appended with values.
// Call .Freeze() method to turn this into proper error.
// Or call .Log() or .Notify() to implicitly freeze the error and do actual stuffs.
//
// Example:
//
//	if err != nil {
//	  return tower.Wrap(err).Message("something went wrong").Freeze()
//	}
//
// Example with Log:
//
//	if err != nil {
//	  return tower.Wrap(err).Message("something went wrong").Log(ctx)
//	}
//
// Example with Notify:
//
//	if err != nil {
//	  return tower.Wrap(err).Message("something went wrong").Notify(ctx)
//	}
//
// Example with Notify and Log:
//
//	if err != nil {
//	  return tower.Wrap(err).Message("something went wrong").Log(ctx).Notify(ctx)
//	}
func Wrap(err error) ErrorBuilder {
	return export.Wrap(err)
}

// Bail creates a new ErrorBuilder from simple messages.
//
// If args are not empty, msg will be fed into fmt.Errorf along with the args.
// Otherwise, msg will be fed into `errors.New()`.
func Bail(msg string, args ...any) ErrorBuilder {
	return export.Bail(msg, args...)
}

// WrapFreeze is a Shorthand for `tower.Wrap(err).Message(message, args...).Freeze()`
//
// Useful when just wanting to add extra simple messages to the error chain.
func WrapFreeze(err error, message string, args ...any) Error {
	return export.WrapFreeze(err, message, args...)
}

// BailFreeze creates new immutable Error from simple messages.
//
// It's a shorthand for `tower.Bail(msg, args...).Freeze()`.
func BailFreeze(msg string, args ...any) Error {
	return export.BailFreeze(msg, args...)
}

// NewEntry Creates a new EntryBuilder. The returned EntryBuilder may be appended with values.
func NewEntry(msg string) EntryBuilder {
	return export.NewEntry(msg)
}

type global int

// Global Methods to handle global instance of Tower.
const Global global = 0

var export *Tower

func init() {
	export = NewTower(Service{})
	export.SetCallerDepth(3)
}

// SetGlobal Set the global tower instance.
// There is no mutex locking, so do not call this method where data race is likely to happen.
func (global) SetGlobal(t *Tower) {
	export = t
}

// Tower Returns the global tower instance.
func (global) Tower() *Tower {
	return export
}
