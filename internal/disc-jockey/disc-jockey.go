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

func NewDiscJockey() *DiscJockey {
	return &DiscJockey{
		mixer:     beep.Mixer{},
		lock:      sync.Mutex{},
		trackList: map[string]*Track{},
	}
}

// Add a new track to the mixtable, the onEnd callback is called when the track is finished
// A new track is played automatically
func (dj *DiscJockey) Add(id string, s beep.StreamSeekCloser, format beep.Format, opt AddTrackOpt) error {

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
		go func(id string) {
			err := dj.Remove(id)
			if err != nil {
				slog.Debug(fmt.Sprintf("[Disc Jockey] :: Error while removing track %s  in callback. Song has been stopped beforehand : %v", id, err))
			}

			if opt.OnEnd != nil {
				opt.OnEnd(id)
			}
		}(id)
	})

	track := &Track{
		Origin: s,
		Decorated: &effects.Volume{
			Streamer: &beep.Ctrl{Streamer: beep.Seq(target, afterPlayCb), Paused: false},
			// Logarithmic base for volume control
			Base: 10,
			// 0 means "not changed from the original stream"
			// InitVolumeDb is in Db, so we need it back to plain log10, hence the /20
			Volume: opt.InitVolumeDb / 20,
			// A track can't be silent using logarithmic volume control
			// so we use a flag instead
			Silent: false,
		},
	}
	dj.lock.Lock()
	defer dj.lock.Unlock()
	dj.trackList[id] = track
	dj.mixer.Add(track.Decorated)
	return nil
}

func (dj *DiscJockey) Remove(id string) error {
	dj.lock.Lock()
	defer dj.lock.Unlock()
	track, err := dj.getTrack(id)
	if err != nil {
		return err
	}
	err = track.Origin.Close()
	if err != nil {
		// This is not a fatal error, we can still remove the track from the list
		// this can happen if during the time the callback was executing, the source
		// autoclosed
		slog.Warn(fmt.Sprintf("while closing track %s: %s", id, err.Error()))
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

func (dj *DiscJockey) ChangeVolume(id string, volumeDeltaDb float64) error {
	dj.lock.Lock()
	defer dj.lock.Unlock()
	track, err := dj.getTrack(id)
	if err != nil {
		return err
	}
	//track.Decorated.Volume = 20 * math.Log10(volumeDeltaDb)
	// On this one, we want to increase **perceived** sound level.
	// We could either choose to increase the Sound power or the sound pressure
	// Human perception is more sensitive to sound pressure, so we're going with that
	// https://en.wikipedia.org/wiki/Decibel
	// Now, sound pressure delta = 10^(volumeDeltaDb/20)
	// Base of the volume filter is 10, so we need to divide by 20
	track.Decorated.Volume += volumeDeltaDb / 20

	// By convention, we consider that a track is silent if its volume is below -60dB
	track.Decorated.Silent = track.Decorated.Volume*20 <= -60

	return nil
}

func (dj *DiscJockey) getTrack(id string) (*Track, error) {
	track, ok := dj.trackList[id]
	if !ok {
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
