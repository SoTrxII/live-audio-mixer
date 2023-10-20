package disc_jockey

import (
	"github.com/faiface/beep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	test_utils "live-audio-mixer/test-utils"
	"testing"
	"time"
)

func TestDiscJockey_AddNoDup(t *testing.T) {
	dj := NewDiscJockey()
	err := dj.Add("test", nil, beep.Format{}, AddTrackOpt{})
	assert.NoError(t, err)
	err = dj.Add("test", nil, beep.Format{}, AddTrackOpt{})
	assert.Error(t, err)
}

func TestDiscJockey_EndCallback(t *testing.T) {
	dj := NewDiscJockey()
	// And write them to a file
	format := beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}
	quack := test_utils.OpenMp3Resource(t, test_utils.Mp3_Quack)
	quackFormat := beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}
	done := make(chan bool, 1)
	err := dj.Add("quack", quack, quackFormat, AddTrackOpt{OnEnd: func(id string) {
		done <- true
	}})
	assert.NoError(t, err)
	// Pull 5 seconds worth of samples. This is more than enough to trigger the callback
	test_utils.GetSamples(t, dj, format.SampleRate.N(time.Second*5))
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "callback not called")
	}
}

// This can happen when the format couldn't be determined
func TestDiscJockey_SampleRateZero(t *testing.T) {
	dj := NewDiscJockey()
	// And write them to a file
	format := beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}
	quack := test_utils.OpenMp3Resource(t, test_utils.Mp3_Quack)
	done := make(chan bool, 1)
	err := dj.Add("quack", quack, beep.Format{}, AddTrackOpt{OnEnd: func(id string) {
		done <- true
	}})
	assert.NoError(t, err)
	// Pull 5 seconds worth of samples. This is more than enough to trigger the callback
	test_utils.GetSamples(t, dj, format.SampleRate.N(time.Second*5))
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "callback not called")
	}
}

func TestDiscJockey_Remove(t *testing.T) {
	dj := NewDiscJockey()
	mockStream := MockStreamer{}
	mockStream.On("Close").Return(nil)
	err := dj.Add("test", &mockStream, beep.Format{}, AddTrackOpt{})
	assert.NoError(t, err)
	err = dj.Remove("test")
	assert.NoError(t, err)
	// If close wasn't called at this point it will panic
	mockStream.AssertExpectations(t)
	err = dj.Remove("test")
	assert.Error(t, err)
}

func TestDiscJockey_SetPaused(t *testing.T) {
	dj := NewDiscJockey()
	mockStream := MockStreamer{}
	mockStream.On("Close").Return(nil)
	err := dj.Add("test", &mockStream, beep.Format{}, AddTrackOpt{})
	assert.NoError(t, err)
	err = dj.SetPaused("test", true)
	assert.NoError(t, err)
	err = dj.SetPaused("test", false)
	assert.NoError(t, err)
	err = dj.SetPaused("test", false)
	assert.NoError(t, err)
	err = dj.Remove("test")
	assert.NoError(t, err)
	// If close wasn't called at this point it will panic
	mockStream.AssertExpectations(t)
	err = dj.SetPaused("test", true)
	assert.Error(t, err)
}

func TestDiscJockey_ChangeVolume(t *testing.T) {
	dj := NewDiscJockey()
	mockStream := MockStreamer{}
	err := dj.Add("test", &mockStream, beep.Format{}, AddTrackOpt{})
	assert.NoError(t, err)
	track, err := dj.getTrack("test")
	err = dj.ChangeVolume("test", -30)
	assert.NoError(t, err)
	assert.Equal(t, -30.0, 20*track.Decorated.Volume)
	assert.Equal(t, false, track.Decorated.Silent)
	err = dj.ChangeVolume("test", 0)
	assert.Equal(t, -30.0, 20*track.Decorated.Volume)
	assert.Equal(t, false, track.Decorated.Silent)
	err = dj.ChangeVolume("test", -30)
	assert.Equal(t, -60.0, 20*track.Decorated.Volume)
	assert.Equal(t, true, track.Decorated.Silent)
	assert.NoError(t, err)
}

type MockStreamer struct {
	mock.Mock
}

// Implement all methods of beep.StreamSeekCloser
func (m *MockStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	args := m.Called(samples)
	return args.Int(0), args.Bool(1)
}
func (m *MockStreamer) Err() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockStreamer) Len() int {
	args := m.Called()
	return args.Int(0)
}
func (m *MockStreamer) Position() int {
	args := m.Called()
	return args.Int(0)
}
func (m *MockStreamer) Seek(p int) error {
	args := m.Called(p)
	return args.Error(0)
}
func (m *MockStreamer) Close() error {
	args := m.Called()
	return args.Error(0)
}
