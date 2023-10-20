package disc_jockey

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"sync"
)

type Track struct {
	// The original stream
	Origin beep.StreamSeekCloser
	// The actual played stream with all required controls
	Decorated *effects.Volume
}

// DiscJockey is a mixer that can play multiple tracks at the same time
type DiscJockey struct {
	mixer     beep.Mixer
	lock      sync.Mutex
	trackList map[string]*Track
}

type AddTrackOpt struct {
	// The initial volume of the track in decibels
	InitVolumeDb float64
	// The callback to call when the track is finished
	OnEnd func(string)
}
