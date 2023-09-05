package rt_wav_encoder

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/wav"
	"github.com/stretchr/testify/assert"
	"live-audio-mixer/test-utils"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"testing"
	"time"
)

const (
	Target                = "mix.wav"
	RecordingDurationSecs = 10
)

type testAssets struct {
	Bg      beep.StreamSeekCloser
	Quack   beep.StreamSeekCloser
	Chicken beep.StreamSeekCloser
	Target  *os.File
	Dir     string
}

func setup(t *testing.T) testAssets {
	dir, err := os.MkdirTemp("", "rt-wav-encoder")
	log.Printf("Temp dir: %s", dir)
	if err != nil {
		t.Fatal(err)
	}
	fMix, err := os.Create(path.Join(dir, Target))
	if err != nil {
		log.Fatal(err)
	}
	return testAssets{
		Bg:      test_utils.OpenMp3Resource(t, test_utils.BgMusic),
		Quack:   test_utils.OpenMp3Resource(t, test_utils.Quack),
		Chicken: test_utils.OpenMp3Resource(t, test_utils.Chicken),
		Target:  fMix,
		Dir:     dir,
	}
}

func teardown(t *testing.T, assets testAssets) {
	err := assets.Bg.Close()
	if err != nil {
		t.Log(err)
	}
	err = assets.Quack.Close()
	if err != nil {
		t.Log(err)
	}
	err = assets.Chicken.Close()
	if err != nil {
		t.Log(err)
	}
	err = assets.Target.Close()
	if err != nil {
		t.Log(err)
	}
	err = os.RemoveAll(assets.Dir)
	if err != nil {
		t.Log(err)
	}
}

// Test with static assets without any changes
func Test_Static(t *testing.T) {
	as := setup(t)
	defer teardown(t, as)
	// Mix them together in a single stream
	mixer := beep.Mixer{}
	mixer.Add(as.Bg, as.Quack)

	// And write them to a file
	mixFormat := beep.Format{SampleRate: 44100, NumChannels: 2, Precision: 2}

	// As the mixed stream emits silence when no stream are playing
	// the stream will not stop on its own
	// So we way for some time before stopping it manually
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT)
	go func(stop chan os.Signal) {
		for {
			select {
			case <-time.After(RecordingDurationSecs * time.Second):
				signalChan <- syscall.SIGINT
			}
		}
	}(signalChan)
	err := Encode(as.Target, &mixer, mixFormat, signalChan)
	if err != nil {
		log.Fatal(err)
	}
	err = as.Target.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Finally, we check that the file duration is
	// the expected one, and that the wav file is correct
	fMix2, err := os.Open(path.Join(as.Dir, Target))
	if err != nil {
		log.Fatal(err)
	}
	mixed, _, err := wav.Decode(fMix2)
	if err != nil {
		log.Fatal(err)
	}
	assert.InDelta(t, RecordingDurationSecs, mixed.Len()/int(mixFormat.SampleRate), 1)
	err = fMix2.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// Switching song in the middle of the recording
func Test_DynamicSwitch(t *testing.T) {
	as := setup(t)
	defer teardown(t, as)

	// Mix them together in a single stream
	mixer := beep.Mixer{}
	mixer.Add(as.Bg, as.Quack)

	// And write them to a file
	mixFormat := beep.Format{SampleRate: 44100, NumChannels: 2, Precision: 2}
	// As the mixed stream emits silence when no stream are playing
	// the stream will not stop on its own
	// So we way for some time before stopping it manually
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT)
	go func() {
		for {
			select {
			case <-time.After(RecordingDurationSecs / 2 * time.Second):
				err := as.Bg.Close()
				if err != nil {
					log.Fatal(err)
				}
				mixer.Add(as.Chicken)
				return
			}
		}
	}()
	go func(stop chan os.Signal) {
		for {
			select {
			case <-time.After(RecordingDurationSecs * time.Second):
				signalChan <- syscall.SIGINT
			}
		}
	}(signalChan)
	err := Encode(as.Target, &mixer, mixFormat, signalChan)
	if err != nil {
		log.Fatal(err)
	}
	// Finally, we check that the file duration is
	// the expected one, and that the wav file is correct
	fMix2, err := os.Open(path.Join(as.Dir, Target))
	if err != nil {
		log.Fatal(err)
	}
	mixed, _, err := wav.Decode(fMix2)
	if err != nil {
		log.Fatal(err)
	}
	assert.InDelta(t, RecordingDurationSecs, mixed.Len()/int(mixFormat.SampleRate), 1)
	err = fMix2.Close()
	if err != nil {
		log.Fatal(err)
	}
}
