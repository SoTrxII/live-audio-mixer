package internal

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/wav"
	"github.com/stretchr/testify/assert"
	disc_jockey "live-audio-mixer/internal/disc-jockey"
	"live-audio-mixer/internal/rt-encoder"
	test_utils "live-audio-mixer/test-utils"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// This test suite is meant to test a live manipulation/recording of audio

func TestLive_MixedTracks(t *testing.T) {
	dj := disc_jockey.NewDiscJockey()
	chicken := test_utils.OpenMp3Resource(t, test_utils.Mp3_Chicken)
	bg := test_utils.OpenMp3Resource(t, test_utils.Mp3_BgMusic)
	quack := test_utils.OpenMp3Resource(t, test_utils.Mp3_Quack)
	defer chicken.Close()
	defer bg.Close()

	// And write them to a file
	format := beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}
	err := dj.Add("chicken", chicken, beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}, disc_jockey.AddTrackOpt{})
	assert.NoError(t, err)
	//speaker.Play(dj)
	done := make(chan os.Signal, 1)

	go func(done chan os.Signal) {
		select {
		case <-time.After(5 * time.Second):
			err = dj.Remove("chicken")
			assert.NoError(t, err)
			err = dj.Add("bg", bg, beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}, disc_jockey.AddTrackOpt{})
			assert.NoError(t, err)
			err = dj.Add("quack", quack, beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}, disc_jockey.AddTrackOpt{})
			assert.NoError(t, err)

		}

		select {
		case <-time.After(5 * time.Second):
			done <- syscall.SIGINT
		}
	}(done)

	tmpDir, err := os.MkdirTemp("", "lam-live-recording")
	fmt.Println("Temp dir is : " + tmpDir)
	assert.NoError(t, err)
	const testFileName = "mix-candidate"
	oggFile := filepath.Join(tmpDir, testFileName+".ogg")
	wavFile := filepath.Join(tmpDir, testFileName+".wav")
	testFile, err := os.Create(oggFile)
	if err != nil {
		t.Fatal(err)
	}
	err = rt_encoder.FFEncode(testFile, dj, format, done)
	if err != nil {
		t.Fatal(err)
	}

	dj.CloseAll()
	testFile.Close()
	err = test_utils.ToWav(oggFile, wavFile)

	assert.NoError(t, err)

	original := test_utils.OpenWavResource(t, test_utils.Wav_Mix1)
	defer original.Close()
	originalSamples := test_utils.GetAllSamples(t, original)

	f, err := os.Open(wavFile)
	assert.NoError(t, err)

	candidate, _, err := wav.Decode(f)
	defer func() {
		candidate.Close()
		err := os.RemoveAll(tmpDir)
		if err != nil {
			fmt.Println("Warn :: Could not remove temp dir: ", err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	candidatesSamples := test_utils.GetAllSamples(t, candidate)
	sim := test_utils.GetSimilarity(originalSamples, candidatesSamples)
	assert.InDelta(t, 1, sim, 0.1)
}
