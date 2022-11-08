package tower

var exported *Tower

func init() {
	exported = NewTower(Service{})
}

// Returns the global tower instance.
func GetTower() *Tower {
	return exported
}

// Wraps the error. The returned ErrorBuilder may be appended with values.
// Then calls .Freeze() method to turn this into proper error.
func Wrap(err error) ErrorBuilder {
	return exported.Wrap(err)
}

func NewEntry(msg string) Entry {
	panic("implement me")
}

// Set the global tower instance.
// There is no mutex locking, so do not call this method where data race is likely to happen.
func SetGlobal(t *Tower) {
	exported = t
}
