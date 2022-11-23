package towerdiscord

type HighlightHint interface {
	DiscordHighlight() string
}

type MimetypeHint interface {
	Mimetype() string
}
