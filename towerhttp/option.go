package towerhttp

type option struct{}

// Option holds all the available options for responding with towerhttp.
var Option option

func (option) Respond() OptionRespondGroup {
	return OptionRespondGroup{}
}

func (option) RespondHook() RespondHookOptionGroup {
	return RespondHookOptionGroup{}
}

func (option) ClientHook() ClientHookOptionBuilder {
	return ClientHookOptionBuilder{}
}
