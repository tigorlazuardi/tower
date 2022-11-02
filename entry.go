package tower

import "context"

type EntryBuilder interface {
	/*
		Sets the code for this entry.
	*/
	Code(i int) EntryBuilder

	/*
		Sets the message for this entry.

		In built in implementation, If args are supplied, fmt.Sprintf will be called with s as base string.

		Very unlikely you will need to set this, because tower already create the message field from `tower.Info` or it's derivative.
	*/
	Message(s string, args ...any) EntryBuilder

	/*
		Sets additional data that will enrich how the entry will look.

		`tower.Fields` is a type that is well integrated with built in Messengers.
		Using this type as Context value will have the performance more optimized when being marshaled
		or provides additional quality of life improvements without having to implement those features
		yourself. Use `tower.F` as alias for this type.

		In built-in implementation, additional call to .Context() will make additional index, not replacing what you already set.

		Example:

			tower.Wrap(err).Code(400).Context(tower.F{"foo": "bar"}).Freeze()
	*/
	Context(ctx interface{}) EntryBuilder

	/*
		Sets the key for this entry. This is how Messenger will use to identify if an entry is the same as previous or not.

		In tower's built-in implementation, by default, no key is set when creating new entry.

		Usually by not setting the key, The Messenger will generate their own key for this message.

		In built in implementation, If args are supplied, fmt.Sprintf will be called with key as base string.
	*/
	Key(key string, args ...any) EntryBuilder

	/*
		Sets the caller for this entry.

		In tower's built-in implementation, by default, the caller is the location where you call `tower.NewEntry`.
	*/
	Caller(c Caller) EntryBuilder

	/*
		Sets the level for this entry.

		In tower's built-in implementation, this defaults to what method you call to generate this entry.
	*/
	Level(lvl Level) EntryBuilder

	/*
		Freeze this entry. Preventing further mutations.
	*/
	Freeze() Entry
	/*
		Log this entry. Implicitly calling .Freeze() method.
	*/
	Log(ctx context.Context) Entry

	/*
		Sends this Entry to Messengers. Implicitly calling .Freeze() method.
	*/
	Notify(ctx context.Context, opts ...MessageOption) Entry
}

type Entry interface {
	CodeHint
	HTTPCodeHint
	MessageHint
	CallerHint
	ContextHint
	LevelHint
	ErrorUnwrapper
	ErrorWriter

	/*
		Logs this error.
	*/
	Log(ctx context.Context) Error
	/*
		Notifies this error to Messengers.
	*/
	Notify(ctx context.Context, opts ...MessageOption) Error
}
