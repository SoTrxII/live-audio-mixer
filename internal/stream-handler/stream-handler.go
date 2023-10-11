package stream_handler

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/gabriel-vasile/mimetype"
	"log/slog"
	"net/http"
	"strings"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

// GetStream takes an audio URL and returns a beep stream, format and error
func (h *Handler) GetStream(audioUrl string) (beep.StreamSeekCloser, beep.Format, error) {
	// Fetch the audio file from the URL
	resp, err := http.Get(audioUrl)
	if err != nil {
		fmt.Println("Error fetching audio:", err)
		return nil, beep.Format{}, err
	}

	// Note, we don't need to close the stream directly, the Close method of the decoder will do it for us
	// Determine the audio format based on the response content type
	var format beep.Format
	var decoder beep.StreamSeekCloser
	contentType := h.getMimeType(resp)
	if !strings.HasPrefix(contentType, "audio/") {
		slog.Warn(fmt.Sprintf("[Stream handler] :: Invalid content type: '%s' for audio with url %s. Aborting playback", contentType, audioUrl))
		return nil, beep.Format{}, fmt.Errorf("invalid content type: '%s' for audio with url %s. Aborting playback", contentType, audioUrl)
	}

	if contentType == "audio/mpeg" {
		decoder, format, err = mp3.Decode(resp.Body)
		if err == nil {
			return decoder, format, nil
		}
		slog.Warn(fmt.Sprintf("while decoding mp3: %s. Using default decoder", err.Error()))
	}

	sc := NewStreamConverter(audioUrl)
	format = beep.Format{SampleRate: 48000, NumChannels: 2, Precision: 2}
	pipe, err := sc.GetOutput()
	if err != nil {
		return nil, beep.Format{}, err
	}
	go watchEncoder(audioUrl, sc)
	return flac.Decode(pipe)
}

func (h *Handler) getMimeType(res *http.Response) string {
	contentType := res.Header.Get("Content-Type")
	// If the content type is not set in the request or somehow wrong, we double check it
	if contentType == "" || !strings.HasPrefix(contentType, "audio/") {
		mType, err := mimetype.DetectReader(res.Body)
		if err != nil {
			slog.Warn(fmt.Sprintf("[Stream handler] :: Error while detecting mime type of %s: %v", res.Request.URL, err))
			return ""
		}
		contentType = mType.String()
	}
	return contentType
}

func watchEncoder(url string, sc *StreamConverter) {
	errCh := make(chan error)
	go sc.Start(errCh)
	err := <-errCh
	if err != nil {
		// A broken pipe is most likely the player closing the created pipe, it's not that important
		if strings.HasSuffix(url, ".flac") || strings.Contains(err.Error(), "Broken pipe") {
			slog.Debug(fmt.Sprintf("[Stream handler] :: Broken pipe for url %s while encoding stream: %v", url, err))
		} else {
			slog.Warn(fmt.Sprintf("[Stream handler] :: Error for url %s while encoding stream: %v", url, err))
		}
	}
}
