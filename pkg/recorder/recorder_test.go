/**
 * @file recorder_test.go
 * @brief Test for recorder, everything else is mocked
 */
package recorder

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"os"
	"testing"
	"time"
)

func TestRecorder_StartCatchingErrors(t *testing.T) {
	mockEncoder := &mockEncoder{}
	mockEncoder.On("Encode").Return(fmt.Errorf("Test"))
	rec := NewRecorder(nil, mockEncoder.Encode)
	errCh := rec.Start(nil)
	select {
	case err := <-errCh:
		assert.Error(t, err)
	case <-time.After(5 * time.Second):
		assert.Fail(t, "timeout")
	}
}

func TestRecorder_StartAndStop(t *testing.T) {
	mockEncoder := &mockEncoder{}
	mockEncoder.On("Encode").Return(nil)
	rec := NewRecorder(nil, mockEncoder.Encode)
	errCh := rec.Start(nil)
	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		assert.Fail(t, "timeout")
	}
	rec.Stop()
}

type mockEncoder struct {
	mock.Mock
}

func (m *mockEncoder) Encode(w io.WriteSeeker, s beep.Streamer, format beep.Format, signalCh chan os.Signal) (err error) {
	args := m.Called()
	return args.Error(0)
}

type fileStreamer struct {
}

func (fs *fileStreamer) GetStream(string) (beep.StreamSeekCloser, beep.Format, error) {
	return nil, beep.Format{}, nil
}
