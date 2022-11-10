package tower

// Wraps the error. The returned ErrorBuilder may be appended with values.
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

// Shorthand for `tower.Wrap(err).Message(message).Freeze()`
//
// Useful when just wanting to add extra simple messages to the error chain.
func WrapFreeze(err error, message string) Error {
	return export.WrapFreeze(err, message)
}

// Creates a new EntryBuilder. The returned EntryBuilder may be appended with values.
func NewEntry(msg string) EntryBuilder {
	return export.NewEntry(msg)
}

type global int

// Methods to handle global instance of Tower.
const Global global = 0

var export *Tower

func init() {
	export = NewTower(Service{})
}

// Set the global tower instance.
// There is no mutex locking, so do not call this method where data race is likely to happen.
func (global) SetGlobal(t *Tower) {
	export = t
}

// Returns the global tower instance.
func (global) Tower() *Tower {
	return export
}
