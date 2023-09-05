package internal

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/wav"
	"github.com/stretchr/testify/assert"
	disc_jockey "live-audio-mixer/internal/disc-jockey"
	rt_wav_encoder "live-audio-mixer/internal/rt-wav-encoder"
	stream_handler "live-audio-mixer/internal/stream-handler"
	test_utils "live-audio-mixer/test-utils"
	"os"
	"syscall"
	"testing"
	"time"
)

// This test suite is meant to test a live manipulation/recording of audio

func TestLive_MixedTracks(t *testing.T) {
	dj := disc_jockey.NewDiscJockey()
	chicken := test_utils.OpenMp3Resource(t, test_utils.Chicken)
	bg := test_utils.OpenMp3Resource(t, test_utils.BgMusic)
	quack := test_utils.OpenMp3Resource(t, test_utils.Quack)
	defer chicken.Close()
	defer bg.Close()

	// And write them to a file
	format := beep.Format{SampleRate: 44100, NumChannels: 2, Precision: 2}
	err := dj.Add("chicken", chicken, beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}, nil)
	assert.NoError(t, err)
	//speaker.Play(dj)
	done := make(chan os.Signal, 1)

	go func(done chan os.Signal) {
		select {
		case <-time.After(5 * time.Second):
			err = dj.Remove("chicken")
			assert.NoError(t, err)
			err = dj.Add("bg", bg, beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}, nil)
			assert.NoError(t, err)
			err = dj.Add("quack", quack, beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}, nil)
			assert.NoError(t, err)

		}

		select {
		case <-time.After(5 * time.Second):
			done <- syscall.SIGINT
		}
	}(done)

	const testFileName = "mix-candidate.wav"
	testFile, err := os.Create(testFileName)
	if err != nil {
		t.Fatal(err)
	}
	err = rt_wav_encoder.Encode(testFile, dj, format, done)
	if err != nil {
		t.Fatal(err)
	}

	dj.CloseAll()
	testFile.Close()
	original := test_utils.OpenWavResource(t, test_utils.Mix1)
	defer original.Close()
	originalSamples := test_utils.GetAllSamples(t, original)

	f, err := os.Open(testFileName)
	if err != nil {
		t.Fatal(err)
	}
	candidate, _, err := wav.Decode(f)
	defer func() {
		candidate.Close()
		err := os.Remove(testFileName)
		if err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	candidatesSamples := test_utils.GetAllSamples(t, candidate)
	sim := test_utils.GetSimilarity(originalSamples, candidatesSamples)
	assert.InDelta(t, 1, sim, 0.1)
}

func TestLive_FromURLs(t *testing.T) {
	dj := disc_jockey.NewDiscJockey()
	// And write them to a file
	format := beep.Format{SampleRate: 44100, NumChannels: 2, Precision: 2}
	h := stream_handler.NewHandler()

	stream, format, err := h.GetStream("https://download.samplelib.com/mp3/sample-3s.mp3")
	err = dj.Add("chicken", stream, format, nil)
	assert.NoError(t, err)
	//speaker.Play(dj)
	done := make(chan os.Signal, 1)

	go func(done chan os.Signal) {
		/*select {
		case <-time.After(5 * time.Second):
			err = dj.Remove("chicken")
			assert.NoError(t, err)
			err = dj.Add("bg", bg, beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}, nil)
			assert.NoError(t, err)
			err = dj.Add("quack", quack, beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}, nil)
			assert.NoError(t, err)

		}*/

		select {
		case <-time.After(5 * time.Second):
			done <- syscall.SIGINT
		}
	}(done)

	const testFileName = "mix-candidate.wav"
	testFile, err := os.Create(testFileName)
	if err != nil {
		t.Fatal(err)
	}
	err = rt_wav_encoder.Encode(testFile, dj, format, done)
	if err != nil {
		t.Fatal(err)
	}

	dj.CloseAll()
	testFile.Close()
	/*original := test_utils.OpenWavResource(t, test_utils.Mix1)
	defer original.Close()
	originalSamples := getSamples(t, original)

	f, err := os.Open(testFileName)
	if err != nil {
		t.Fatal(err)
	}
	candidate, _, err := wav.Decode(f)
	defer func() {
		candidate.Close()
		err := os.Remove(testFileName)
		if err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	candidatesSamples := getSamples(t, candidate)
	sim := test_utils.GetSimilarity(originalSamples, candidatesSamples)
	assert.InDelta(t, 1, sim, 0.1)*/

}
