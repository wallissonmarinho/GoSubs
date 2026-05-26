package domain

type SubtitleTokenPayload struct {
	Media      MediaRef `json:"media"`
	SourceID   string   `json:"source_id"`
	SourceURL  string   `json:"source_url"`
	SourceLang string   `json:"source_lang"`
	Format     string   `json:"format"`
	Encoding   string   `json:"encoding"`
	Release    string   `json:"release,omitempty"`
	TargetLang string   `json:"target_lang"`
}

type CleanupResult struct {
	Scanned    int    `json:"scanned"`
	Removed    int    `json:"removed"`
	FreedBytes int64  `json:"freed_bytes"`
	TTL        string `json:"ttl"`
}
