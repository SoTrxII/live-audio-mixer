package test_utils

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/wav"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type Resource string

const (
	BgMusic         Resource = "Sappheiros_Falling.mp3"
	Quack                    = "quack.mp3"
	Chicken                  = "chicken_song.mp3"
	Castle                   = "barovian-castle.mp3"
	Sample3s                 = "sample-3s.mp3"
	BabyElephant             = "baby-elephant.wav"
	Ensoniq                  = "ensoniq.wav"
	Rec_NoLoop               = "./recorder/no-loop.wav"
	Rec_Loop                 = "./recorder/loop.wav"
	Rec_StartStop            = "./recorder/start-stop.wav"
	Rec_MultiTracks          = "./recorder/multi-tracks.wav"
	Mix1                     = "./mixed/mix1.wav"
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
