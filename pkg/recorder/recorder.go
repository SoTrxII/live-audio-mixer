package recorder

import (
	"fmt"
	"github.com/faiface/beep"
	disc_jockey "live-audio-mixer/internal/disc-jockey"
	pb "live-audio-mixer/proto"
	"log/slog"
	"os"
	"sync"
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
		err = r.addTrack(evt.AssetUrl)
	case pb.EventType_STOP:
		err = r.removeTrack(evt.AssetUrl)
	case pb.EventType_PAUSE:
		err = r.pauseTrack(evt.AssetUrl)
	case pb.EventType_RESUME:
		err = r.resumeTrack(evt.AssetUrl)
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
		err := r.addTrack(url)
		if err != nil {
			return err
		}
	}
	return nil
}

// Add a track to the mixtable from its URL
func (r *Recorder) addTrack(url string) error {
	stream, format, err := r.src.GetStream(url)
	if err != nil {
		return err
	}
	err = r.dj.Add(url, stream, format, func() {
		err := r.loop(url)
		if err != nil {
			slog.Error(fmt.Sprintf("[Recorder] :: Error while looping track %s : %v", url, err))
		}
	})
	//err = r.dj.Add(url, stream, format, nil)
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
