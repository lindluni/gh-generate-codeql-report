// Package codeql provides functionality to interact with GitHub CodeQL API.
package codeql

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/go-github/v72/github"
)

// Alert represents processed CodeQL alert data.
type Alert struct {
	Owner       string
	Repo        string
	ID          int
	Severity    string
	ShortDesc   string
	FullDesc    string
	FilePath    string
	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int
}

// Client handles interactions with GitHub's CodeQL API.
type Client struct {
	ghClient *github.Client
	logger   *log.Logger
	lastRate *github.Rate
}

// NewClient creates a new CodeQL client with the provided token.
func NewClient(token string, logger *log.Logger) *Client {
	return &Client{
		ghClient: github.NewClient(nil).WithAuthToken(token),
		logger:   logger,
	}
}

// GetAlert fetches a CodeQL alert by its number.
func (c *Client) GetAlert(ctx context.Context, owner, repo string, alertNumber int64) (*Alert, error) {
	c.logger.Printf("Fetching alert #%d for %s/%s", alertNumber, owner, repo)

	for {
		alert, resp, err := c.ghClient.CodeScanning.GetAlert(ctx, owner, repo, alertNumber)
		if err != nil {
			// Check for rate limit error
			if resp != nil && resp.StatusCode == http.StatusForbidden {
				if rl := resp.Rate; rl.Remaining == 0 {
					reset := rl.Reset.Time.Sub(time.Now())
					if reset > 0 {
						c.logger.Printf("GitHub rate limit reached. Sleeping for %v until %v", reset, rl.Reset.Time)
						time.Sleep(reset)
						continue // retry after sleep
					}
				}
			}
			return nil, fmt.Errorf("failed to get alert: %w", err)
		}

		// Log and store rate limit info
		if resp != nil {
			c.lastRate = &resp.Rate
			if c.lastRate.Remaining < 10 {
				c.logger.Printf("Warning: GitHub API rate limit low: %d remaining, resets at %v", c.lastRate.Remaining, c.lastRate.Reset.Time)
			}
		}

		location := alert.MostRecentInstance.GetLocation()
		return &Alert{
			Owner:       owner,
			Repo:        repo,
			ID:          alert.GetNumber(),
			Severity:    alert.Rule.GetSecuritySeverityLevel(),
			ShortDesc:   alert.Rule.GetDescription(),
			FullDesc:    alert.Rule.GetFullDescription(),
			FilePath:    location.GetPath(),
			StartLine:   location.GetStartLine(),
			StartColumn: location.GetStartColumn(),
			EndLine:     location.GetEndLine(),
			EndColumn:   location.GetEndColumn(),
		}, nil
	}
}
