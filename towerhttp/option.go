package towerhttp

type option struct{}

// Option holds all the available options for responding with towerhttp.
var Option option

func (option) Respond() RespondOptionBuilder {
	return RespondOptionBuilder{}
}

func (option) RespondHook() RespondHookOptionGroup {
	return RespondHookOptionGroup{}
}

func (option) RoundTripHook() RoundTripHookOptionBuilder {
	return RoundTripHookOptionBuilder{}
}

func (option) RoundTrip() RoundTripOptionBuilder {
	return RoundTripOptionBuilder{}
}
