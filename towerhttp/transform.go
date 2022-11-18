package towerhttp

import "context"

type BodyTransform interface {
	// BodyTransform transform given input into another shape. This is called before the Encoder.
	BodyTransform(ctx context.Context, input any) any
}

// BodyTransformFunc is a convenient function that implements BodyTransform.
type BodyTransformFunc func(ctx context.Context, input any) any

func (b BodyTransformFunc) BodyTransform(ctx context.Context, input any) any {
	return b(ctx, input)
}

// NoopBodyTransform is a BodyTransform that does nothing and only return the input as is.
type NoopBodyTransform struct{}

func (n NoopBodyTransform) BodyTransform(ctx context.Context, input any) any {
	return input
}
