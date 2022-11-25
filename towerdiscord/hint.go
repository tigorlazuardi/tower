package towerdiscord

type HighlightHint interface {
	DiscordHighlight() string
}

type MimetypeHint interface {
	Mimetype() string
}

type ImageSizeHint interface {
	ImageSize() (width, height int)
}
