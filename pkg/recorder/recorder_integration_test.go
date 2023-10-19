/**
 * @file recorder_integration_test.go
 * @brief Integration test for recorder, we're using the real encoder and stream handler
 */
package recorder

import (
	"fmt"
	"github.com/faiface/beep/wav"
	"github.com/stretchr/testify/assert"
	rt_encoder "live-audio-mixer/internal/rt-encoder"
	stream_handler "live-audio-mixer/internal/stream-handler"
	pb "live-audio-mixer/proto"
	test_utils "live-audio-mixer/test-utils"
	"os"
	"path/filepath"
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
	fLoc := path + "/rec.ogg"
	fmt.Println(fLoc)
	file, err := os.Create(fLoc)
	assert.NoError(t, err)
	return &testSetup{
		Rec:  NewRecorder(stream_handler.NewHandler(), rt_encoder.FFEncode),
		Dir:  path,
		File: file,
		destroy: func(t *testing.T) {
			err := file.Close()
			if err != nil {
				fmt.Println("Warn :: Could not remove file")
			}
			err = os.RemoveAll(path)
			if err != nil {
				fmt.Println("Warn :: Could not remove tmpdir")

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

	// To be able to compare the file, we must encode the ogg file to wav
	wavPath := filepath.Join(test.Dir, "rec.wav")
	err := test_utils.ToWav(test.File.Name(), wavPath)
	assert.NoError(t, err)

	// Then we can test both files
	wavFile, err := os.Open(wavPath)
	defer wavFile.Close()
	assert.NoError(t, err)
	candidate, _, err := wav.Decode(wavFile)
	original := test_utils.OpenWavResource(t, test_utils.Wav_Rec_StartStop)
	sim := test_utils.GetSimilarity(test_utils.GetAllSamples(t, original), test_utils.GetAllSamples(t, candidate))

	// We can't expect 100% similarity, because the encoder is not lossless
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

	// To be able to compare the file, we must encode the ogg file to wav
	wavPath := filepath.Join(test.Dir, "rec.wav")
	err := test_utils.ToWav(test.File.Name(), wavPath)
	assert.NoError(t, err)

	// Then we can test both files
	wavFile, err := os.Open(wavPath)
	defer wavFile.Close()
	assert.NoError(t, err)
	candidate, _, err := wav.Decode(wavFile)
	original := test_utils.OpenWavResource(t, test_utils.Wav_Rec_NoLoop)
	sim := test_utils.GetSimilarity(test_utils.GetAllSamples(t, original), test_utils.GetAllSamples(t, candidate))

	// We can't expect 100% similarity, because the encoder is not lossless
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

	// To be able to compare the file, we must encode the ogg file to wav
	wavPath := filepath.Join(test.Dir, "rec.wav")
	err := test_utils.ToWav(test.File.Name(), wavPath)
	assert.NoError(t, err)

	// Then we can test both files
	wavFile, err := os.Open(wavPath)
	defer wavFile.Close()
	assert.NoError(t, err)
	candidate, _, err := wav.Decode(wavFile)
	original := test_utils.OpenWavResource(t, test_utils.Wav_Rec_Loop)
	sim := test_utils.GetSimilarity(test_utils.GetAllSamples(t, original), test_utils.GetAllSamples(t, candidate))

	// We can't expect 100% similarity, because the encoder is not lossless
	assert.InDelta(t, 1, sim, 0.15)
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

	// To be able to compare the file, we must encode the ogg file to wav
	wavPath := filepath.Join(test.Dir, "rec.wav")
	err := test_utils.ToWav(test.File.Name(), wavPath)
	assert.NoError(t, err)

	// Then we can test both files
	wavFile, err := os.Open(wavPath)
	defer wavFile.Close()
	assert.NoError(t, err)
	candidate, _, err := wav.Decode(wavFile)
	original := test_utils.OpenWavResource(t, test_utils.Wav_Rec_MultiTracks)
	sim := test_utils.GetSimilarity(test_utils.GetAllSamples(t, original), test_utils.GetAllSamples(t, candidate))

	// We can't expect 100% similarity, because the encoder is not lossless
	// This one also depend on the resampling
	assert.InDelta(t, 1, sim, 0.2)
}

func TestPauseResume(t *testing.T) {
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
		time.Sleep(5 * time.Second)
		test.Rec.Update(&pb.Event{
			Type:     pb.EventType_PAUSE,
			AssetUrl: "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			Loop:     true,
		})
		time.Sleep(3 * time.Second)
		test.Rec.Update(&pb.Event{
			Type:     pb.EventType_RESUME,
			AssetUrl: "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			Loop:     true,
		})
		time.Sleep(5 * time.Second)
		test.Rec.Stop()
	}()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	}

	// To be able to compare the file, we must encode the ogg file to wav
	wavPath := filepath.Join(test.Dir, "rec.wav")
	err := test_utils.ToWav(test.File.Name(), wavPath)
	assert.NoError(t, err)

	// Then we can test both files
	wavFile, err := os.Open(wavPath)
	defer wavFile.Close()
	assert.NoError(t, err)
	candidate, _, err := wav.Decode(wavFile)
	original := test_utils.OpenWavResource(t, test_utils.Wav_Rec_PauseResume)
	sim := test_utils.GetSimilarity(test_utils.GetAllSamples(t, original), test_utils.GetAllSamples(t, candidate))

	// We can't expect 100% similarity, because the encoder is not lossless
	// This one also depend on the resampling
	assert.InDelta(t, 1, sim, 0.2)
}

func TestVolume(t *testing.T) {
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

		// Halve perceived volume
		test.Rec.Update(&pb.Event{
			Type:          pb.EventType_VOLUME,
			AssetUrl:      "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			Loop:          true,
			VolumeDeltaDb: -3,
		})
		time.Sleep(5 * time.Second)

		// This should trigger the mute state
		test.Rec.Update(&pb.Event{
			Type:          pb.EventType_VOLUME,
			AssetUrl:      "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			Loop:          true,
			VolumeDeltaDb: -57,
		})
		time.Sleep(5 * time.Second)

		// Back to normal
		test.Rec.Update(&pb.Event{
			Type:          pb.EventType_VOLUME,
			AssetUrl:      "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			Loop:          true,
			VolumeDeltaDb: 60,
		})
		time.Sleep(5 * time.Second)
		test.Rec.Stop()
	}()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	}
	// TODO add comparison
}

func TestSeek(t *testing.T) {
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

		// Halve perceived volume
		test.Rec.Update(&pb.Event{
			Type:            pb.EventType_SEEK,
			AssetUrl:        "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			SeekPositionSec: 20,
		})
		time.Sleep(5 * time.Second)

		// This should trigger the mute state
		test.Rec.Update(&pb.Event{
			Type:            pb.EventType_SEEK,
			AssetUrl:        "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			SeekPositionSec: 0,
		})
		time.Sleep(5 * time.Second)
		test.Rec.Stop()
	}()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	}

	// To be able to compare the file, we must encode the ogg file to wav
	wavPath := filepath.Join(test.Dir, "rec.wav")
	err := test_utils.ToWav(test.File.Name(), wavPath)
	assert.NoError(t, err)

	// Then we can test both files
	wavFile, err := os.Open(wavPath)
	defer wavFile.Close()
	assert.NoError(t, err)
	candidate, _, err := wav.Decode(wavFile)
	original := test_utils.OpenWavResource(t, test_utils.Wav_Rec_Seek)
	sim := test_utils.GetSimilarity(test_utils.GetAllSamples(t, original), test_utils.GetAllSamples(t, candidate))

	// We can't expect 100% similarity, because the encoder is not lossless
	// This one also depend on the resampling
	assert.InDelta(t, 1, sim, 0.1)
}
