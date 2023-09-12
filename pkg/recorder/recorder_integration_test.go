/**
 * @file recorder_integration_test.go
 * @brief Integration test for recorder, we're using the real encoder and stream handler
 */
package recorder

import (
	"fmt"
	"github.com/faiface/beep/wav"
	"github.com/stretchr/testify/assert"
	"io"
	rt_wav_encoder "live-audio-mixer/internal/rt-wav-encoder"
	stream_handler "live-audio-mixer/internal/stream-handler"
	pb "live-audio-mixer/proto"
	test_utils "live-audio-mixer/test-utils"
	"os"
	"testing"
	"time"
)

type testSetup struct {
	Rec     *Recorder
	Dir     string
	File    *os.File
	destroy func(t *testing.T)
}

func setup(t *testing.T) *testSetup {
	path, err := os.MkdirTemp("", "test-recorder")
	assert.NoError(t, err)
	fLoc := path + "/rec.wav"
	fmt.Println(fLoc)
	file, err := os.Create(fLoc)
	assert.NoError(t, err)
	return &testSetup{
		Rec:  NewRecorder(stream_handler.NewHandler(), rt_wav_encoder.Encode),
		Dir:  path,
		File: file,
		destroy: func(t *testing.T) {
			err := file.Close()
			if err != nil {
				t.Fatal(err)
			}
			err = os.RemoveAll(path)
			if err != nil {
				t.Fatal(err)
			}
		},
	}
}

func TestStartStop(t *testing.T) {
	test := setup(t)
	defer test.destroy(t)
	errCh := test.Rec.Start(test.File)
	go func() {
		time.Sleep(1 * time.Second)
		test.Rec.Update(&pb.Event{
			Type:     pb.EventType_PLAY,
			AssetUrl: "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			Loop:     false,
		})
		time.Sleep(5 * time.Second)
		test.Rec.Update(&pb.Event{
			Type:     pb.EventType_STOP,
			AssetUrl: "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			Loop:     false,
		})
		time.Sleep(5 * time.Second)
		test.Rec.Stop()
	}()
	select {
	case err := <-errCh:
		assert.NoError(t, err)
	}
	_, err := test.File.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	candidate, _, err := wav.Decode(test.File)
	original := test_utils.OpenWavResource(t, test_utils.Wav_Rec_StartStop)
	sim := test_utils.GetSimilarity(test_utils.GetAllSamples(t, original), test_utils.GetAllSamples(t, candidate))
	assert.InDelta(t, 1, sim, 0.1)
}

func TestNoLoop(t *testing.T) {
	test := setup(t)
	defer test.destroy(t)
	errCh := test.Rec.Start(test.File)
	go func() {
		time.Sleep(1 * time.Second)
		test.Rec.Update(&pb.Event{
			Type:     pb.EventType_PLAY,
			AssetUrl: "https://download.samplelib.com/mp3/sample-3s.mp3",
			Loop:     false,
		})
		time.Sleep(9 * time.Second)
		test.Rec.Stop()
	}()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	}

	_, err := test.File.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	candidate, _, err := wav.Decode(test.File)
	original := test_utils.OpenWavResource(t, test_utils.Wav_Rec_NoLoop)
	sim := test_utils.GetSimilarity(test_utils.GetAllSamples(t, original), test_utils.GetAllSamples(t, candidate))
	assert.InDelta(t, 1, sim, 0.1)
}

func TestLoop(t *testing.T) {
	test := setup(t)
	defer test.destroy(t)
	errCh := test.Rec.Start(test.File)
	go func() {
		time.Sleep(1 * time.Second)
		test.Rec.Update(&pb.Event{
			Type:     pb.EventType_PLAY,
			AssetUrl: "https://download.samplelib.com/mp3/sample-3s.mp3",
			Loop:     true,
		})
		time.Sleep(9 * time.Second)
		test.Rec.Stop()
	}()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	}

	_, err := test.File.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	candidate, _, err := wav.Decode(test.File)
	original := test_utils.OpenWavResource(t, test_utils.Wav_Rec_Loop)
	sim := test_utils.GetSimilarity(test_utils.GetAllSamples(t, original), test_utils.GetAllSamples(t, candidate))
	assert.InDelta(t, 1, sim, 0.1)
}

func TestMultiTrack(t *testing.T) {
	test := setup(t)
	defer test.destroy(t)
	errCh := test.Rec.Start(test.File)
	go func() {
		time.Sleep(1 * time.Second)
		test.Rec.Update(&pb.Event{
			Type:     pb.EventType_PLAY,
			AssetUrl: "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			Loop:     false,
		})
		test.Rec.Update(&pb.Event{
			Type:     pb.EventType_PLAY,
			AssetUrl: "https://download.samplelib.com/mp3/sample-3s.mp3",
			Loop:     true,
		})
		time.Sleep(9 * time.Second)
		test.Rec.Update(&pb.Event{
			Type:     pb.EventType_STOP,
			AssetUrl: "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			Loop:     false,
		})
		time.Sleep(5 * time.Second)
		test.Rec.Stop()
	}()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	}

	_, err := test.File.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	candidate, _, err := wav.Decode(test.File)
	original := test_utils.OpenWavResource(t, test_utils.Wav_Rec_MultiTracks)
	sim := test_utils.GetSimilarity(test_utils.GetAllSamples(t, original), test_utils.GetAllSamples(t, candidate))
	assert.InDelta(t, 1, sim, 0.1)
}
