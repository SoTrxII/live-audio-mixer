package stream_handler

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"log/slog"
	"net/http"
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
	switch resp.Header.Get("Content-Type") {
	case "audio/mpeg":
		decoder, format, err = mp3.Decode(resp.Body)
		if err != nil {
			wrappedErr := fmt.Errorf("while decoding mp3: %w", err)
			return nil, beep.Format{}, wrappedErr
		}
		// TODO: Add wav support
	/*case "audio/wav", "audio/x-wav":
	decoder, format, err = wav.Decode(resp.Body)
	if err != nil {
		wrappedErr := fmt.Errorf("while decoding WAV: %w", err)
		return nil, beep.Format{}, wrappedErr
	}*/
	case "":
		slog.Info("[Stream handler] :: No content type header found. Defaulting to mpeg")
		decoder, format, err = mp3.Decode(resp.Body)
		if err != nil {
			wrappedErr := fmt.Errorf("while decoding mp3: %w", err)
			return nil, beep.Format{}, wrappedErr
		}
	default:
		return nil, beep.Format{}, fmt.Errorf("Unsupported audio format %s\n", resp.Header.Get("Content-Type"))
	}
	return decoder, format, nil
}
