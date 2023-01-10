package towerhttp

type ClientHookOption interface {
	apply(*clientHook)
}

type ClientHookOptionBuilder []ClientHookOption

func (c ClientHookOptionBuilder) apply(hook *clientHook) {
	for _, opt := range c {
		opt.apply(hook)
	}
}

func (c ClientHookOptionBuilder) ReadRequestBodyLimit(limit int) ClientHookOptionBuilder {
	return append(c, ClientHookOptionFunc(func(hook *clientHook) {
		hook.readRequestLimit = limit
	}))
}

func (c ClientHookOptionBuilder) ReadResponseBodyLimit(limit int) ClientHookOptionBuilder {
	return append(c, ClientHookOptionFunc(func(hook *clientHook) {
		hook.readRespondLimit = limit
	}))
}

func (c ClientHookOptionBuilder) FilterRequest(filter FilterRequest) ClientHookOptionBuilder {
	return append(c, ClientHookOptionFunc(func(hook *clientHook) {
		hook.filterRequest = filter
	}))
}

func (c ClientHookOptionBuilder) FilterResponse(filter FilterClientResponse) ClientHookOptionBuilder {
	return append(c, ClientHookOptionFunc(func(hook *clientHook) {
		hook.filterResponse = filter
	}))
}

func (c ClientHookOptionBuilder) Log(log ClientExecuteHookFunc) ClientHookOptionBuilder {
	return append(c, ClientHookOptionFunc(func(hook *clientHook) {
		hook.log = log
	}))
}

type ClientHookOptionFunc func(*clientHook)

func (c ClientHookOptionFunc) apply(hook *clientHook) {
	c(hook)
}

func NewClientHook(opts ...ClientHookOption) ClientHook {
	c := &clientHook{}
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}
