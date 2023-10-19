package recorder

import (
	"github.com/faiface/beep"
	"io"
	disc_jockey "live-audio-mixer/internal/disc-jockey"
	pb "live-audio-mixer/proto"
	"os"
	"sync"
	"time"
)

type Recorder struct {
	dj    *disc_jockey.DiscJockey
	src   StreamingSrc
	state map[string]*pb.Event
	sink  Sink
	mu    sync.Mutex
}

type EncodeFn func(w io.WriteSeeker, s beep.Streamer, format beep.Format, signalCh chan os.Signal) (err error)
type Sink struct {
	fn   func(w io.WriteSeeker, s beep.Streamer, format beep.Format, signalCh chan os.Signal) (err error)
	stop chan os.Signal
	ack  chan error
}

type StreamingSrc interface {
	GetStream(string, time.Duration) (beep.StreamSeekCloser, beep.Format, error)
}
