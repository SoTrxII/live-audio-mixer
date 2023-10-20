package recorder

import (
	"fmt"
	"github.com/faiface/beep"
	disc_jockey "live-audio-mixer/internal/disc-jockey"
	pb "live-audio-mixer/proto"
	"log/slog"
	"os"
	"sync"
	"time"
)

func NewRecorder(src StreamingSrc, to EncodeFn) *Recorder {
	return &Recorder{
		dj:    disc_jockey.NewDiscJockey(),
		state: map[string]*pb.Event{},
		src:   src,
		sink:  Sink{fn: to, stop: make(chan os.Signal, 1), ack: make(chan error, 1)},
		mu:    sync.Mutex{},
	}
}

func (r *Recorder) Start(to *os.File) chan error {
	// Starts encoding asynchronously
	go func(stop chan os.Signal, ack chan error) {
		format := beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}
		ack <- r.sink.fn(to, r.dj, format, stop)
	}(r.sink.stop, r.sink.ack)

	return r.sink.ack
}

func (r *Recorder) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sink.stop <- os.Interrupt
}

func (r *Recorder) Update(evt *pb.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var err error
	r.state[evt.AssetUrl] = evt
	switch evt.Type {
	case pb.EventType_PLAY:
		err = r.addTrack(evt.AssetUrl, evt.VolumeDeltaDb, 0)
	case pb.EventType_STOP:
		err = r.removeTrack(evt.AssetUrl)
	case pb.EventType_PAUSE:
		err = r.pauseTrack(evt.AssetUrl)
	case pb.EventType_RESUME:
		err = r.resumeTrack(evt.AssetUrl)
	case pb.EventType_VOLUME:
		err = r.changeVolume(evt.AssetUrl, evt.VolumeDeltaDb)
	case pb.EventType_SEEK:
		err = r.seekTrack(evt.AssetUrl, evt.VolumeDeltaDb, time.Duration(evt.SeekPositionSec)*time.Second)
	// This type of event only toggles the loop flag currently, there is no processing required
	case pb.EventType_OTHER:
		slog.Info(fmt.Sprintf("[Recorder] :: Received OTHER event %v", evt))
	default:
		slog.Warn(fmt.Sprintf("[Recorder] :: Unknown event type %v", evt.Type))
	}
	if err != nil {
		slog.Error(fmt.Sprintf("[Recorder] :: Error while handling event %v : %v", evt, err))
	}
}

func (r *Recorder) loop(url string) error {
	lastEvt := r.state[url]
	if lastEvt.Loop {
		err := r.addTrack(url, lastEvt.VolumeDeltaDb, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

// Add a track to the mixtable from its URL
func (r *Recorder) addTrack(url string, initVolume float64, offset time.Duration) error {
	stream, format, err := r.src.GetStream(url, offset)
	if err != nil {
		return err
	}
	err = r.dj.Add(url, stream, format, disc_jockey.AddTrackOpt{
		InitVolumeDb: initVolume,
		OnEnd: func(url string) {
			err := r.loop(url)
			if err != nil {
				slog.Error(fmt.Sprintf("[Recorder] :: Error while looping track %s : %v", url, err))
			}
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// Remove a track from the mixtable
func (r *Recorder) removeTrack(url string) error {
	return r.dj.Remove(url)
}

// Pause a track
func (r *Recorder) pauseTrack(url string) error {
	return r.dj.SetPaused(url, true)
}

// Resume a track
func (r *Recorder) resumeTrack(url string) error {
	return r.dj.SetPaused(url, false)
}

func (r *Recorder) changeVolume(url string, volumeDeltaDb float64) error {
	return r.dj.ChangeVolume(url, volumeDeltaDb)
}

func (r *Recorder) seekTrack(url string, initVolume float64, offset time.Duration) error {
	err := r.removeTrack(url)
	if err != nil {
		return err
	}
	return r.addTrack(url, initVolume, offset)
}
