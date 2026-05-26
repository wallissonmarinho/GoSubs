package domain

type SubtitleCandidate struct {
	ID                 string
	URL                string
	Format             string
	Encoding           string
	Display            string
	Language           string
	Source             string
	Release            string
	FileName           string
	MatchedRelease     string
	MatchedFilter      string
	AI                 bool
	IsHearingImpaired  bool
}

type SubtitleSearchRequest struct {
	Media       MediaRef
	ReleaseHint string
}

type ListedSubtitle struct {
	ID   string `json:"id,omitempty"`
	URL  string `json:"url"`
	Lang string `json:"lang"`
}

type SubtitleListResponse struct {
	Subtitles []ListedSubtitle `json:"subtitles"`
}

type SubtitleTextChunk struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
