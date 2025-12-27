package converter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/kkdai/youtube/v2"
	"go.leapkit.dev/core/server"
)

func Convert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		server.Errorf(w, http.StatusBadRequest, "failed to parse form: %w", err)
		return
	}

	videoURL := r.FormValue("youtube-url")
	if videoURL == "" {
		server.Errorf(w, http.StatusBadRequest, "youtube-url is required")
		return
	}

	// Get video info first
	client := youtube.Client{}
	video, err := client.GetVideo(videoURL)
	if err != nil {
		server.Errorf(w, http.StatusInternalServerError, "failed to get video info: %w", err)
		return
	}

	sanitizedTitle := sanitizeFilename(video.Title)

	// Set headers before starting conversion
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.mp3\"", sanitizedTitle))
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Cache-Control", "no-cache")

	// Stream directly to response writer
	if err := streamYouTubeToMP3(r.Context(), video, w); err != nil {
		fmt.Fprintf(os.Stderr, "streaming error: %v\n", err)
		return
	}
}

// streamYouTubeToMP3 streams the conversion directly to the writer
func streamYouTubeToMP3(ctx context.Context, video *youtube.Video, w io.Writer) error {
	client := youtube.Client{}

	formats := video.Formats.Type("audio")
	if len(formats) == 0 {
		formats = video.Formats.WithAudioChannels()
	}

	if len(formats) == 0 {
		return fmt.Errorf("no audio formats available")
	}

	var smallestFormat *youtube.Format
	for i := range formats {
		f := &formats[i]
		if f.QualityLabel == "" { // Audio-only
			if smallestFormat == nil || f.Bitrate < smallestFormat.Bitrate {
				smallestFormat = f
			}
		}
	}

	if smallestFormat == nil {
		smallestFormat = &formats[0]
	}

	stream, _, err := client.GetStream(video, smallestFormat)
	if err != nil {
		return fmt.Errorf("failed to get video stream: %w", err)
	}
	defer stream.Close()

	// Create ffmpeg command with context for cancellation
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", "pipe:0", // Read from stdin
		"-vn",
		"-ar", "22050", // Lower sample rate
		"-ac", "1", // Mono
		"-b:a", "64k", // Lower bitrate
		"-f", "mp3", // Output format
		"pipe:1", // Write to stdout
	)

	cmd.Stdin = stream
	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w", err)
	}

	return nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "_")
	}
	return name
}
