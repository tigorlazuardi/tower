package tower

import "context"

type EntryBuilder interface {
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

type Entry interface{}
