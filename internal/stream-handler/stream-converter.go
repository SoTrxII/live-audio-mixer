package stream_handler

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type StreamConverter struct {
	cmd    *exec.Cmd
	stderr io.ReadCloser
}

func NewStreamConverter(url string) *StreamConverter {
	return &StreamConverter{
		// Flac
		cmd: exec.Command("ffmpeg", "-ss", "0", "-i", url, "-vn", "-ac", "2", "-ar", "48000", "-acodec", "flac", "-f", "flac", "-"),
		// Wav (Non functional, reason unknown)
		//cmd: exec.Command("ffmpeg", "-i", url, "-vn", "-acodec", "pcm_s16le", "-ar", "48000", "-ac", "2", "-f", "wav", "-frames:v", "48000", "-"),
		// Mp3 (Non functional, requires seeking)
		//cmd: exec.Command("ffmpeg", "-i", url, "-vn", "-ar", "48000", "-acodec", "mp3", "-f", "mp3", "-"),
		// Raw (Produces cracks in sound, probably due to some incorrect rounding on my side)
		//cmd: exec.Command("ffmpeg", "-i", url, "-acodec", "pcm_s16le", "-ar", "48000", "-ac", "2", "-f", "s16le", "-"),
	}

}

func (s *StreamConverter) GetOutput() (pipe *bufio.Reader, err error) {
	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error capturing FFmpeg output:", err)
		return nil, err
	}

	s.stderr, err = s.cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error capturing FFmpeg output:", err)
		return
	}

	return bufio.NewReader(stdout), nil
}

// Starts encoding asynchronously and sends any errors to the error channel
func (s *StreamConverter) Start(errCh chan error) {

	strCh := make(chan string)
	go s.readError(strCh)
	if err := s.cmd.Start(); err != nil {
		errMessage := <-strCh
		errCh <- fmt.Errorf(errMessage)
		return
	}
	if err := s.cmd.Wait(); err != nil {
		errMessage := <-strCh
		errCh <- fmt.Errorf(errMessage)
		return
	}
	errCh <- nil // Send nil to the error channel to indicate that the process has ended
}

func (s *StreamConverter) readError(strCh chan string) {
	buf := make([]byte, 4096)
	var sb strings.Builder
	for {
		n, err := s.stderr.Read(buf)
		if err != nil && err != io.EOF {
			break
		}
		if n > 0 {
			sb.Write(buf[:n])
		}
	}
	strCh <- sb.String()
}
