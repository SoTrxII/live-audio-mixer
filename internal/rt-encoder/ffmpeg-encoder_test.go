package rt_encoder

import (
	"github.com/faiface/beep"
	"github.com/stretchr/testify/assert"
	"live-audio-mixer/test-utils"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	Target                = "mix.ogg"
	RecordingDurationSecs = 15
)

type testAssets struct {
	Bg      beep.StreamSeekCloser
	Quack   beep.StreamSeekCloser
	Chicken beep.StreamSeekCloser
	Target  *os.File
	Dir     string
}

func setup(t *testing.T) testAssets {
	dir, err := os.MkdirTemp("", "rt-encoder")
	log.Printf("Temp dir: %s", dir)
	if err != nil {
		t.Fatal(err)
	}
	fMix, err := os.Create(path.Join(dir, Target))
	if err != nil {
		log.Fatal(err)
	}
	return testAssets{
		Bg:      test_utils.OpenMp3Resource(t, test_utils.Mp3_BgMusic),
		Quack:   test_utils.OpenMp3Resource(t, test_utils.Mp3_Quack),
		Chicken: test_utils.OpenMp3Resource(t, test_utils.Mp3_Chicken),
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
	mixFormat := beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}

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
	err := FFEncode(as.Target, &mixer, mixFormat, signalChan)
	assert.NoError(t, err)
	err = as.Target.Close()
	assert.NoError(t, err)

	// Finally, we check that the file duration is
	// the expected one, and that the wav file is correct
	dur, err := getDuration(path.Join(as.Dir, Target))
	assert.NoError(t, err)
	assert.InDelta(t, RecordingDurationSecs, dur.Seconds(), 1)
}

// Switching song in the middle of the recording
func Test_DynamicSwitch(t *testing.T) {
	as := setup(t)
	defer teardown(t, as)

	// Mix them together in a single stream
	mixer := beep.Mixer{}
	mixer.Add(as.Bg, as.Quack)

	// And write them to a file
	mixFormat := beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}
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
				// Chicken will appear slowed down. This is normal, as the sample rate is 48k, but the mixer is 48K
				// This isn't what we're testing for here
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
	err := FFEncode(as.Target, &mixer, mixFormat, signalChan)
	if err != nil {
		log.Fatal(err)
	}
	// Finally, we check that the file duration is
	// the expected one, and that the wav file is correct
	dur, err := getDuration(path.Join(as.Dir, Target))
	assert.NoError(t, err)
	assert.InDelta(t, RecordingDurationSecs, dur.Seconds(), 1)
}
func getDuration(path string) (time.Duration, error) {
	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 input.mp4
	arg := strings.Split("-v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1", " ")
	arg = append(arg, path)
	cmd := exec.Command("ffprobe", arg...)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	dur, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(int(dur)) * time.Second, nil
}
