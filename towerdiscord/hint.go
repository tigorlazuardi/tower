package towerdiscord

type HighlightHint interface {
	DiscordHighlight() string
}

type MimetypeHint interface {
	Mimetype() string
}

type LengthHint interface {
	Len() int
}

type ImageSizeHint interface {
	ImageSize() (width, height int)
}
