package stream_handler

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/stretchr/testify/assert"
	test_utils "live-audio-mixer/test-utils"
	"testing"
	"time"
)

func setup(t *testing.T) []getStreamCase {
	return []getStreamCase{
		{
			link:       "https://s3.amazonaws.com/cdn.roll20.net/ttaudio/148_Barovian_Castle.mp3",
			compareFor: 5 * time.Second,
			compareTo:  test_utils.OpenMp3Resource(t, test_utils.Castle),
		},
		{
			link:       "https://download.samplelib.com/mp3/sample-3s.mp3",
			compareFor: 3 * time.Second,
			compareTo:  test_utils.OpenMp3Resource(t, test_utils.Sample3s),
		},
		/*{
			link:       "https://www2.cs.uic.edu/~i101/SoundFiles/BabyElephantWalk60.wav",
			compareFor: 20 * time.Second,
			compareTo:  test_utils.OpenWavResource(t, test_utils.BabyElephant),
		},{
			link:       "https://freewavesamples.com/files/Ensoniq-ZR-76-08-Dope-92.wav",
			compareFor: 2 * time.Second,
			compareTo:  test_utils.OpenWavResource(t, test_utils.Ensoniq),
		},*/
	}
}
func TestGetStream(t *testing.T) {
	h := NewHandler()
	cases := setup(t)
	for _, testCase := range cases {
		testName := fmt.Sprintf("Testing valid link %s", testCase.link)
		t.Run(testName, func(t *testing.T) {
			decoder, format, err := h.GetStream(testCase.link)
			assert.NoError(t, err)
			assert.NotNil(t, decoder)
			fmt.Printf("Format: %+v\n", format)
		})
	}

	invalidLinks := []string{"https://file-examples.com/wp-content/storage/2017/11/file_example_OOG_1MG.ogg"}

	for _, link := range invalidLinks {
		testName := fmt.Sprintf("Testing invalid link %s", link)
		t.Run(testName, func(t *testing.T) {
			_, _, err := h.GetStream(link)
			assert.Error(t, err)
		})
	}
}

type getStreamCase struct {
	link       string
	compareFor time.Duration
	compareTo  beep.StreamSeekCloser
}

// Retrieve a stream and compare the content to another one
func TestGetStream_Cmp(t *testing.T) {
	cases := setup(t)
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("Testing link %s", testCase.link), func(t *testing.T) {
			h := NewHandler()

			// Open the candidate stream and get the required number of sample
			candidate, format, err := h.GetStream(testCase.link)
			assert.NoError(t, err)
			candidateSamples := test_utils.GetSamples(t, candidate, format.SampleRate.N(testCase.compareFor))

			// Same thing for the original we're comparing it to
			originalSamples := test_utils.GetSamples(t, testCase.compareTo, format.SampleRate.N(testCase.compareFor))

			// Compare. Each case should be 1.0 but there could be small variations
			simScore := test_utils.GetSimilarity(originalSamples, candidateSamples)
			assert.InDelta(t, 1, simScore, 0.1)
		})
	}

}

// This is a human test, hearing the actual file
func TestGetStream_Listen(t *testing.T) {
	cases := setup(t)
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("Testing link %s", testCase.link), func(t *testing.T) {
			h := NewHandler()
			// Open the candidate stream and get the required number of sample
			candidate, format, err := h.GetStream(testCase.link)
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
