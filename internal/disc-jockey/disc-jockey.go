package disc_jockey

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"log/slog"
	"sync"
)

// A collection of utilities taking a collection of streamseeker and handling control
// related operations like looping, seeking, pausing...

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

func NewDiscJockey() *DiscJockey {
	return &DiscJockey{
		mixer:     beep.Mixer{},
		lock:      sync.Mutex{},
		trackList: map[string]*Track{},
	}
}

// Add a new track to the mixtable, the onEnd callback is called when the track is finished
// A new track is played automatically
func (dj *DiscJockey) Add(id string, s beep.StreamSeekCloser, format beep.Format, onEnd func()) error {

	// abort if the track already exists
	if _, err := dj.getTrack(id); err == nil {
		return fmt.Errorf("track with id %s already exists", id)
	}

	var target beep.Streamer = s
	if format.SampleRate == beep.SampleRate(0) {
		slog.Info(fmt.Sprintf("[Disc Jockey] :: Track %s has a sample rate of 0. Assuming 48000 and hoping for the best", id))
	} else if format.SampleRate != beep.SampleRate(48000) {
		slog.Info(fmt.Sprintf("[Disc Jockey] :: Resampling track %s from %d to 48000", id, format.SampleRate))
		target = beep.Resample(3, format.SampleRate, beep.SampleRate(48000), s)
	}

	// Every time a song stops playing, it is removed from the track list
	afterPlayCb := beep.Callback(func() {
		err := dj.remove(id)
		if err != nil {
			slog.Debug(fmt.Sprintf("[Disc Jockey] :: Error while removing track %s  in callback. Song has been stopped beforehand : %v", id, err))
		}
		if onEnd != nil {
			onEnd()
		}
	})
	track := &Track{
		Origin: s,
		Decorated: &effects.Volume{
			Streamer: &beep.Ctrl{Streamer: beep.Seq(target, afterPlayCb), Paused: false},
			// Logarithmic base for volume control
			Base: 2,
			// 0 means "not changed from the original stream"
			Volume: 0,
			// A track can't be silent using logarithmic volume control
			// so we use a flag instead
			Silent: false,
		},
	}
	// We're going with no lock on this one because of the loop
	//dj.lock.Lock()
	//defer dj.lock.Unlock()
	dj.trackList[id] = track
	dj.mixer.Add(track.Decorated)
	return nil
}

// Remove a track from the mixtable
func (dj *DiscJockey) Remove(id string) error {
	dj.lock.Lock()
	defer dj.lock.Unlock()
	return dj.remove(id)
}
func (dj *DiscJockey) remove(id string) error {
	track, err := dj.getTrack(id)
	if err != nil {
		return err
	}
	err = track.Origin.Close()
	if err != nil {
		return fmt.Errorf("while closing track %s: %w", id, err)
	}
	delete(dj.trackList, id)
	return nil
}

// CloseAll closes all tracks
func (dj *DiscJockey) CloseAll() {
	dj.lock.Lock()
	defer dj.lock.Unlock()
	for _, track := range dj.trackList {
		track.Origin.Close()
	}
	dj.trackList = map[string]*Track{}
}

// SetPaused a single track
func (dj *DiscJockey) SetPaused(id string, paused bool) error {
	dj.lock.Lock()
	defer dj.lock.Unlock()
	track, err := dj.getTrack(id)
	if err != nil {
		return err
	}

	if track.Decorated.Streamer.(*beep.Ctrl).Paused == paused {
		slog.Warn(fmt.Sprintf(`[Disc Jockey] :: Attempting to set track "%s" to state paused=%t, which it is alredy in`, id, paused))
	}
	track.Decorated.Streamer.(*beep.Ctrl).Paused = paused
	return nil
}

func (dj *DiscJockey) getTrack(id string) (*Track, error) {
	track := dj.trackList[id]
	if track == nil {
		return nil, fmt.Errorf(`track with id "%s" not found`, id)
	}
	return track, nil
}

func (dj *DiscJockey) Stream(samples [][2]float64) (n int, ok bool) {
	dj.lock.Lock()
	defer dj.lock.Unlock()
	return dj.mixer.Stream(samples)
}
func (dj *DiscJockey) Err() error {
	return dj.mixer.Err()
}
