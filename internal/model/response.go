package model

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SummaryResponse is the strict JSON shape expected from the model.
type SummaryResponse struct {
	Rows  []SummaryRow `json:"rows"`
	Notes []string     `json:"notes,omitempty"`
}

type SummaryRow struct {
	Area      string `json:"area,omitempty"`
	PathOrURL string `json:"path_or_url"`
	Summary   string `json:"summary"`
	Ticket    string `json:"ticket,omitempty"`
}

func ParseAndValidateJSON(raw string) (*SummaryResponse, error) {
	var sr SummaryResponse
	if err := json.Unmarshal([]byte(raw), &sr); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	if err := sr.Validate(); err != nil {
		return nil, err
	}
	return &sr, nil
}

func (s *SummaryResponse) Validate() error {
	if len(s.Rows) == 0 {
		return errors.New("response must include non-empty \"rows\"")
	}
	for i, r := range s.Rows {
		if r.PathOrURL == "" {
			return fmt.Errorf("rows[%d].path_or_url is required", i)
		}
		if r.Summary == "" {
			return fmt.Errorf("rows[%d].summary is required", i)
		}
	}
	return nil
}
