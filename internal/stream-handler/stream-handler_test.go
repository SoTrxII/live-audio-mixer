package stream_handler

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/stretchr/testify/assert"
	test_utils "live-audio-mixer/test-utils"
	"net/http"
	"os"
	"testing"
	"time"
)

func setup(t *testing.T) []getStreamCase {
	return []getStreamCase{
		{
			link:       "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			offset:     5 * time.Second,
			compareFor: 5 * time.Second,
			compareTo:  test_utils.OpenFlacResource(t, test_utils.Flac_Castle),
			mime:       "audio/mpeg",
		},
		{
			link: "https://download.samplelib.com/mp3/sample-3s.mp3",
			// Negative offset should have no effect
			offset:     -3 * time.Second,
			compareFor: 3 * time.Second,
			compareTo:  test_utils.OpenFlacResource(t, test_utils.Flac_Sample3s),
			mime:       "audio/mpeg",
		},
		{
			link:       "https://youtube.songbroker.pocot.fr/download/mp3/3cFvVYUxwoc-#87867",
			compareFor: 10 * time.Second,
			compareTo:  test_utils.OpenFlacResource(t, test_utils.Flac_mp3_Layer3),
			mime:       "audio/mpeg",
		},
		{
			link:       "https://upload.wikimedia.org/wikipedia/commons/c/c8/Example.ogg",
			compareFor: 5 * time.Second,
			compareTo:  test_utils.OpenFlacResource(t, test_utils.Flac_SampleOpus),
			mime:       "audio/ogg",
		},
		{
			link:       "https://www2.cs.uic.edu/~i101/SoundFiles/BabyElephantWalk60.wav",
			offset:     10 * time.Second,
			compareFor: 20 * time.Second,
			compareTo:  test_utils.OpenFlacResource(t, test_utils.Flac_BabyElephant),
			mime:       "audio/x-wav",
		}, {
			link:       "https://freewavesamples.com/files/Ensoniq-ZR-76-08-Dope-92.wav",
			compareFor: 2 * time.Second,
			compareTo:  test_utils.OpenFlacResource(t, test_utils.Flac_Ensoniq),
			mime:       "audio/wav",
		},
		{
			link:       "https://freewavesamples.com/files/Ensoniq-ZR-76-08-Dope-92.wav",
			compareFor: 10 * time.Second,
			compareTo:  test_utils.OpenFlacResource(t, test_utils.Flac_Ensoniq),
			mime:       "audio/wav",
		},
	}
}
func TestGetStream(t *testing.T) {
	h := NewHandler()
	cases := setup(t)
	for _, testCase := range cases {
		testName := fmt.Sprintf("Testing valid link %s", testCase.link)
		t.Run(testName, func(t *testing.T) {
			decoder, format, err := h.GetStream(testCase.link, 0)
			assert.NoError(t, err)
			assert.NotNil(t, decoder)
			fmt.Printf("Format: %+v\n", format)
		})
	}

	invalidLinkCases := []getStreamCase{
		// Cloudlfare protected link
		{
			link: "https://file-examples.com/wp-content/storage/2017/11/file_example_OOG_1MG.ogg",
		},
	}

	for _, testCase := range invalidLinkCases {
		testName := fmt.Sprintf("Testing invalid link %s", testCase.link)
		t.Run(testName, func(t *testing.T) {
			_, _, err := h.GetStream(testCase.link, 0)
			assert.Error(t, err)
		})
	}
}

type getStreamCase struct {
	link string
	// The offset is used to skip the first N seconds of the stream
	offset     time.Duration
	compareFor time.Duration
	compareTo  beep.StreamSeekCloser
	mime       string
}

// Retrieve a stream and compare the content to another one
func TestGetStream_Cmp(t *testing.T) {
	cases := setup(t)
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("Testing link %s", testCase.link), func(t *testing.T) {
			h := NewHandler()
			// Open the candidate stream and get the required number of sample
			candidate, format, err := h.GetStream(testCase.link, testCase.offset)
			assert.NoError(t, err)
			candidateSamples := test_utils.GetSamples(t, candidate, format.SampleRate.N(testCase.compareFor))

			// Same thing for the original we're comparing it to, ignoring the offset sample
			_ = test_utils.GetSamples(t, testCase.compareTo, format.SampleRate.N(testCase.offset))
			originalSamples := test_utils.GetSamples(t, testCase.compareTo, format.SampleRate.N(testCase.compareFor))

			// Compare. Each case should be 1.0 but there could be small variations
			simScore := test_utils.GetSimilarity(originalSamples, candidateSamples)
			assert.InDelta(t, 1, simScore, 0.1)
		})
	}

}

func TestHandler_GetMimeType(t *testing.T) {
	cases := setup(t)
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("Testing link %s", testCase.link), func(t *testing.T) {
			h := NewHandler()
			res, err := http.Get(testCase.link)
			assert.NoError(t, err)
			mime := h.getMimeType(res)
			assert.Equal(t, testCase.mime, mime)
		})
	}
}

// This is a human test, hearing the actual file
func TestGetStream_Listen(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping human test in GitHub Actions environment")
	}
	cases := setup(t)
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("Testing link %s", testCase.link), func(t *testing.T) {
			h := NewHandler()
			// Open the candidate stream and get the required number of sample
			candidate, format, err := h.GetStream(testCase.link, 10*time.Second)
			assert.NoError(t, err)
			err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			assert.NoError(t, err)
			speaker.Play(candidate)
			select {
			case <-time.After(testCase.compareFor):
				return
			}
		})
	}
}

func TestHandler_GetStream_NotAnAudioFile(t *testing.T) {
	h := NewHandler()
	_, _, err := h.GetStream("https://www.google.com", time.Second)
	assert.Error(t, err)
}

func TestHandler_GetStream_InvalidLink(t *testing.T) {
	h := NewHandler()
	_, _, err := h.GetStream("fhdfhdfhhfd://garbage.com", time.Second)
	assert.Error(t, err)
}
