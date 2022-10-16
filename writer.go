package tower

import "io"

type Writer interface {
	io.Writer
	io.StringWriter
}

type Display interface {
	// Display returns a human readable and rich with information for the implementer.
	Display() string
}

type DisplayWriter interface {
	// Writes the Display() string to the writer instead of being allocated as value.
	WriteDisplay(w Writer)
}

type ErrorWriter interface {
	// Writes the error.Error to the writer instead of being allocated as value.
	WriteError(w Writer)
}

type Summary interface {
	// Returns a short summary of the implementer.
	Summary() string
}

type SummaryWriter interface {
	// Writes the Summary() string to the writer instead of being allocated as value.
	WriteSummary(w Writer)
}
