package stream_handler

import (
	"fmt"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/speaker"
	"github.com/stretchr/testify/assert"
	"io"
	test_utils "live-audio-mixer/test-utils"
	"os"
	"strings"
	"testing"
	"time"
)

var testCases = []string{
	"https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
	"https://freetestdata.com/wp-content/uploads/2021/09/Free_Test_Data_2MB_OGG.ogg",
	"https://filesamples.com/samples/audio/flac/Symphony%20No.6%20(1st%20movement).flac",
}

func TestStreamConverter_StartNoError(t *testing.T) {
	for _, testLink := range testCases {
		t.Run(fmt.Sprintf("Testing valid link %s", testLink), func(t *testing.T) {
			convert := NewStreamConverter(testLink)
			stdout, err := convert.GetOutput()
			assert.NoError(t, err)
			errCh := make(chan error)
			go convert.Start(errCh)
			// Create a buffer to read the output
			buffer := make([]byte, 10*1024*1024) // Adjust the buffer size as needed
			for {
				_, err := stdout.Read(buffer)
				if err == io.EOF {
					break // End of output
				}
				if err != nil {
					fmt.Println("Error reading stdout:", err)
					break
				}
			}
			res := <-errCh
			assert.NoError(t, res)
		})
	}
}

func TestStreamConverter_StartError(t *testing.T) {
	convert := NewStreamConverter("http://garbage.com")
	stdout, err := convert.GetOutput()
	assert.NoError(t, err)
	errCh := make(chan error)
	go convert.Start(errCh)
	// Create a buffer to read the output
	buffer := make([]byte, 10*1024*1024) // Adjust the buffer size as needed

	for {
		_, err := stdout.Read(buffer)
		if err == io.EOF {
			break // End of output
		}
		if err != nil {
			fmt.Println("Error reading stdout:", err)
			break
		}
	}
	res := <-errCh
	fmt.Printf("Error: %+v\n", res)
	assert.Error(t, res)
}

func TestStreamConverter_Listen(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping human test in GitHub Actions environment")
	}
	for _, testLink := range testCases {
		t.Run(fmt.Sprintf("Testing valid link %s", testLink), func(t *testing.T) {
			convert := NewStreamConverter(testLink)
			pipe, err := convert.GetOutput()
			assert.NoError(t, err)
			errCh := make(chan error)
			go convert.Start(errCh)
			stream, format, err := flac.Decode(pipe)
			assert.NoError(t, err)
			assert.NotNil(t, stream)
			err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			assert.NoError(t, err)
			speaker.Play(stream)
			select {
			case <-time.After(5 * time.Second):
				return
			}
		})
	}

}

// Ensures the process is killed from the outbound stream is closed
func TestStreamConverter_Release(t *testing.T) {
	for _, testLink := range testCases {
		t.Run(fmt.Sprintf("Testing valid link %s", testLink), func(t *testing.T) {
			convert := NewStreamConverter(testLink)
			pipe, err := convert.GetOutput()
			assert.NoError(t, err)
			errCh := make(chan error)
			go convert.Start(errCh)
			stream, _, err := flac.Decode(pipe)
			assert.NoError(t, err)
			assert.NotNil(t, stream)
			test_utils.GetSamples(t, stream, 4096)
			err = stream.Close()
			assert.NoError(t, err)
			res := <-errCh

			// If the input stream is closed, it may cause a "Broken pipe" error
			// This is the expected behavior, as allowing the process to continue would cause a memory leak
			// This is however different with flac files, which will not cause a broken pipe error
			// but another one as the file is not "piped"
			if strings.HasSuffix(testLink, ".flac") {
				return
			}
			assert.Contains(t, res.Error(), "Broken pipe")
		})
	}

}
