package domain

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type SubtitleCue struct {
	Index int
	Start string
	End   string
	Text  string
}

func ParseSRT(data []byte) ([]SubtitleCue, error) {
	sc := bufio.NewScanner(bytes.NewReader(normalizeSRTNewlines(data)))
	var cues []SubtitleCue
	for {
		if !sc.Scan() {
			break
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		idx, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("invalid cue index: %w", err)
		}
		if !sc.Scan() {
			return nil, fmt.Errorf("missing timing line")
		}
		timing := strings.TrimSpace(sc.Text())
		parts := strings.Split(timing, " --> ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid timing line")
		}
		var textLines []string
		for sc.Scan() {
			txt := sc.Text()
			if strings.TrimSpace(txt) == "" {
				break
			}
			textLines = append(textLines, txt)
		}
		cues = append(cues, SubtitleCue{
			Index: idx,
			Start: parts[0],
			End:   parts[1],
			Text:  strings.Join(textLines, "\n"),
		})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return cues, nil
}

func BuildSRT(cues []SubtitleCue) []byte {
	var b strings.Builder
	for i, cue := range cues {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%d\n%s --> %s\n%s\n", cue.Index, cue.Start, cue.End, cue.Text)
	}
	return []byte(b.String())
}

func normalizeSRTNewlines(data []byte) []byte {
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
	return data
}
