package domain

import (
	"fmt"
	"strconv"
	"strings"
)

type MediaRef struct {
	Type    string
	IMDbID  string
	TMDBID  int
	Season  int
	Episode int
}

func ParseMediaRef(contentType, rawID, extra string) (MediaRef, error) {
	if v := parseVideoIDFromExtra(extra); v != "" {
		rawID = v
	}
	rawID = strings.TrimSpace(rawID)
	if rawID == "" {
		return MediaRef{}, fmt.Errorf("media id required")
	}
	if strings.HasPrefix(rawID, "tt") {
		return MediaRef{Type: contentType, IMDbID: rawID}, nil
	}
	if strings.HasPrefix(rawID, "tmdb:") {
		parts := strings.Split(rawID, ":")
		if len(parts) == 2 {
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				return MediaRef{}, fmt.Errorf("invalid tmdb movie id")
			}
			return MediaRef{Type: contentType, TMDBID: id}, nil
		}
		if len(parts) == 4 {
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				return MediaRef{}, fmt.Errorf("invalid tmdb series id")
			}
			season, err := strconv.Atoi(parts[2])
			if err != nil {
				return MediaRef{}, fmt.Errorf("invalid season")
			}
			episode, err := strconv.Atoi(parts[3])
			if err != nil {
				return MediaRef{}, fmt.Errorf("invalid episode")
			}
			return MediaRef{Type: contentType, TMDBID: id, Season: season, Episode: episode}, nil
		}
	}
	return MediaRef{}, fmt.Errorf("unsupported media id %q", rawID)
}

func parseVideoIDFromExtra(extra string) string {
	extra = strings.TrimPrefix(strings.TrimSpace(extra), "/")
	if extra == "" {
		return ""
	}
	extra = strings.TrimSuffix(extra, ".json")
	for _, part := range strings.Split(extra, "&") {
		if !strings.HasPrefix(part, "videoID=") {
			continue
		}
		return strings.TrimPrefix(part, "videoID=")
	}
	return ""
}
