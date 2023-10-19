package test_utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/wav"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type Resource string

const (
	Mp3_BgMusic         Resource = "Sappheiros_Falling.mp3"
	Mp3_Quack                    = "quack.mp3"
	Mp3_Chicken                  = "chicken_song.mp3"
	Flac_mp3_Layer3              = "sample-mp3-layer3.flac"
	Flac_Castle                  = "barovian-castle.flac"
	Flac_Sample3s                = "sample-3s.flac"
	Flac_SampleOpus              = "sample-opus.flac"
	Flac_BabyElephant            = "baby-elephant-stereo.flac"
	Flac_Ensoniq                 = "ensoniq.flac"
	Wav_Rec_NoLoop               = "./recorder/no-loop.wav"
	Wav_Rec_Loop                 = "./recorder/loop.wav"
	Wav_Rec_StartStop            = "./recorder/start-stop.wav"
	Wav_Rec_MultiTracks          = "./recorder/multi-tracks.wav"
	Wav_Rec_PauseResume          = "./recorder/pause-resume.wav"
	Wav_Rec_Seek                 = "./recorder/seek.wav"
	Wav_Mix1                     = "./mixed/mix1.wav"
)
const (
	ResPath = "../resources/"
)

func GetResAbsolutePath(t *testing.T, r Resource) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Couldn't get current file path")
	}
	dir := filepath.Dir(filename)
	return filepath.Join(dir, ResPath, string(r))
}

func OpenMp3Resource(t *testing.T, r Resource) beep.StreamSeekCloser {
	f, err := os.Open(GetResAbsolutePath(t, r))
	if err != nil {
		t.Fatal(err)
	}
	decoded, _, err := mp3.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}

func OpenWavResource(t *testing.T, r Resource) beep.StreamSeekCloser {
	f, err := os.Open(GetResAbsolutePath(t, r))
	if err != nil {
		t.Fatal(err)
	}
	decoded, _, err := wav.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}

func OpenFlacResource(t *testing.T, r Resource) beep.StreamSeekCloser {
	f, err := os.Open(GetResAbsolutePath(t, r))
	if err != nil {
		t.Fatal(err)
	}
	decoded, _, err := flac.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}

func ToWav(src, dst string) error {
	cmdLine := fmt.Sprintf("-i %s -y -acodec pcm_s16le -ac 2 -f wav %s", src, dst)
	cmd := exec.Command("ffmpeg", strings.Split(cmdLine, " ")...)
	cmd.Stderr = os.Stderr
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return err
}

// Checksum returns the SHA-256 checksum of the specified file
func GetChecksum(filename string) (string, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// Create a new SHA-256 hasher
	hasher := sha256.New()
	// Copy the file contents to the hasher
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}
	// Get the checksum as a byte slice
	checksumBytes := hasher.Sum(nil)
	// Convert the checksum to a hexadecimal string
	checksum := hex.EncodeToString(checksumBytes)
	return checksum, nil
}
