package process

type Flag int

type ProcessOptions struct {
	Flag Flag
}

type ProcessOption func(opts *ProcessOptions)
