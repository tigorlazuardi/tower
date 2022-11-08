package tower

import (
	"context"
	"fmt"
)

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

			tower.NewEntry(msg).Code(200).Context(tower.F{"foo": "bar"}).Freeze()
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

	/*
		Logs this entry.
	*/
	Log(ctx context.Context) Entry
	/*
		Notifies this entry to Messengers.
	*/
	Notify(ctx context.Context, opts ...MessageOption) Entry
}

type EntryConstructorContext struct {
	Caller  Caller
	Message string
	Tower   *Tower
}

type EntryConstructor interface {
	ConstructEntry(*EntryConstructorContext) EntryBuilder
}

var _ EntryConstructor = (EntryConstructorFunc)(nil)

type EntryConstructorFunc func(*EntryConstructorContext) EntryBuilder

func (f EntryConstructorFunc) ConstructEntry(ctx *EntryConstructorContext) EntryBuilder {
	return f(ctx)
}

func defaultEntryConstructor(ctx *EntryConstructorContext) EntryBuilder {
	return &entryBuilder{
		message: ctx.Message,
		caller:  ctx.Caller,
		context: []any{},
		level:   InfoLevel,
		tower:   ctx.Tower,
	}
}

type entryBuilder struct {
	code    int
	message string
	caller  Caller
	context []any
	key     string
	level   Level
	tower   *Tower
}

func (e *entryBuilder) Code(i int) EntryBuilder {
	e.code = i
	return e
}

func (e *entryBuilder) Message(s string, args ...any) EntryBuilder {
	if len(args) > 0 {
		e.message = fmt.Sprintf(s, args...)
	} else {
		e.message = s
	}
	return e
}

func (e *entryBuilder) Context(ctx interface{}) EntryBuilder {
	e.context = append(e.context, ctx)
	return e
}

func (e *entryBuilder) Key(key string, args ...any) EntryBuilder {
	e.key = key
	return e
}

func (e *entryBuilder) Caller(c Caller) EntryBuilder {
	e.caller = c
	return e
}

func (e *entryBuilder) Level(lvl Level) EntryBuilder {
	e.level = lvl
	return e
}

func (e *entryBuilder) Freeze() Entry {
	return implEntry{e}
}

func (e *entryBuilder) Log(ctx context.Context) Entry {
	return e.Freeze().Log(ctx)
}

func (e *entryBuilder) Notify(ctx context.Context, opts ...MessageOption) Entry {
	return e.Freeze().Notify(ctx, opts...)
}

type implEntry struct {
	inner *entryBuilder
}

// Gets the original code of the type.
func (e implEntry) Code() int {
	return e.inner.code
}

// Gets HTTP Status Code for the type.
func (e implEntry) HTTPCode() int {
	switch {
	case e.inner.code >= 200 && e.inner.code <= 599:
		return e.inner.code
	case e.inner.code > 999:
		code := e.inner.code % 1000
		if code >= 200 && code <= 599 {
			return code
		}
	}
	return 200
}

// Gets the Message of the type.
func (e implEntry) Message() string {
	return e.inner.message
}

// Gets the caller of this type.
func (e implEntry) Caller() Caller {
	return e.inner.caller
}

// Gets the context of this this type.
func (e implEntry) Context() []any {
	return e.inner.context
}

// Gets the level of this message.
func (e implEntry) Level() Level {
	return e.inner.level
}

/*
Logs this entry.
*/
func (e implEntry) Log(ctx context.Context) Entry {
	e.inner.tower.Log(ctx, e)
	return e
}

/*
Notifies this entry to Messengers.
*/
func (e implEntry) Notify(ctx context.Context, opts ...MessageOption) Entry {
	e.inner.tower.Notify(ctx, e, opts...)
	return e
}
