package rt_encoder

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"time"
)

func FFEncode(w io.WriteSeeker, s beep.Streamer, format beep.Format, signalCh chan os.Signal) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "wav")
		}
	}()

	if format.NumChannels <= 0 {
		return errors.New("wav: invalid number of channels (less than 1)")
	}
	if format.Precision != 1 && format.Precision != 2 && format.Precision != 3 {
		return errors.New("wav: unsupported precision, 1, 2 or 3 is supported")
	}
	pipeReader, pipeWriter := io.Pipe()
	cmd := exec.Command("ffmpeg", "-re", "-f", "s16le", "-ar", "44100", "-ac", "2", "-i", "pipe:0", "-c:a", "libopus", "-f", "ogg", "pipe:1")
	cmd.Stdin = pipeReader
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	/*defer func() {
		err := cmd.Wait()
		if err != nil {
			fmt.Println("Error waiting for ffmpeg to finish:", err)
		}
	}()*/

	recordingDuration := 1 * time.Second // Recording duration of one second.
	chunkSize := int(44100) * 1 * int(recordingDuration.Seconds())
	var (
		bw      = pipeWriter
		samples = make([][2]float64, chunkSize)
		buffer  = make([]byte, len(samples)*format.Width())
		written int
	)
Loop:
	for {
		select {
		case <-signalCh:
			// Can't use signals, as they are not compatible on both Windows and Linux
			err = cmd.Process.Kill()
			if err != nil {
				return err
			}
			break Loop
		default:
			n, ok := s.Stream(samples)
			if !ok {
				break
			}
			buf := buffer
			switch {
			case format.Precision == 1:
				for _, sample := range samples[:n] {
					buf = buf[format.EncodeUnsigned(buf, sample):]
				}
			case format.Precision == 2 || format.Precision == 3:
				for _, sample := range samples[:n] {
					buf = buf[format.EncodeSigned(buf, sample):]
				}
			default:
				panic(fmt.Errorf("wav: encode: invalid precision: %d", format.Precision))
			}
			nn, err := bw.Write(buffer[:n*format.Width()])
			if err != nil {
				return err
			}
			written += nn
		}
	}
	return nil
}
