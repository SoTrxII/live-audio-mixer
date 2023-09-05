package rt_wav_encoder

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/pkg/errors"
)

type header struct {
	RiffMark      [4]byte
	FileSize      int32
	WaveMark      [4]byte
	FmtMark       [4]byte
	FormatSize    int32
	FormatType    int16
	NumChans      int16
	SampleRate    int32
	ByteRate      int32
	BytesPerFrame int16
	BitsPerSample int16
	DataMark      [4]byte
	DataSize      int32
}

// Encode writes a beep stream with a custom buffer size to be near real time (1 second).
//
// Format precision must be 1 or 2 bytes.
func Encode(w io.WriteSeeker, s beep.Streamer, format beep.Format, signalCh chan os.Signal) (err error) {
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

	h := header{
		RiffMark:      [4]byte{'R', 'I', 'F', 'F'},
		FileSize:      -1, // finalization
		WaveMark:      [4]byte{'W', 'A', 'V', 'E'},
		FmtMark:       [4]byte{'f', 'm', 't', ' '},
		FormatSize:    16,
		FormatType:    1,
		NumChans:      int16(format.NumChannels),
		SampleRate:    int32(format.SampleRate),
		ByteRate:      int32(int(format.SampleRate) * format.NumChannels * format.Precision),
		BytesPerFrame: int16(format.NumChannels * format.Precision),
		BitsPerSample: int16(format.Precision) * 8,
		DataMark:      [4]byte{'d', 'a', 't', 'a'},
		DataSize:      -1, // finalization
	}
	if err := binary.Write(w, binary.LittleEndian, &h); err != nil {
		return err
	}
	recordingDuration := time.Second // Recording duration of one second.
	chunkSize := int(44100) * 1 * int(recordingDuration.Seconds())
	var (
		bw      = bufio.NewWriter(w)
		samples = make([][2]float64, chunkSize)
		buffer  = make([]byte, len(samples)*format.Width())
		written int
	)
Loop:
	for {
		select {
		case <-time.After(recordingDuration):
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
		case <-signalCh:
			break Loop
		}
	}
	if err := bw.Flush(); err != nil {
		return err
	}

	// finalize header
	h.FileSize = int32(44 + written) // 44 is the size of the header
	h.DataSize = int32(written)
	if _, err := w.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &h); err != nil {
		return err
	}
	if _, err := w.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	return nil
}
