package whatsmeow

import (
	"context"
	"encoding/json"

	"go.mau.fi/whatsmeow/types"
)

// NewsletterDirectoryFilters filters the channel directory query.
type NewsletterDirectoryFilters struct {
	CountryCodes []string `json:"country_codes,omitempty"`
}

// NewsletterDirectoryInput is the input for GetNewsletterDirectory.
type NewsletterDirectoryInput struct {
	View        string                      `json:"view,omitempty"` // known value: "RECOMMENDED"
	Limit       int                         `json:"limit,omitempty"`
	StartCursor string                      `json:"start_cursor,omitempty"`
	Filters     *NewsletterDirectoryFilters `json:"filters,omitempty"`
}

// NewsletterDirectoryPage is a page of channels from the directory/recommended queries.
type NewsletterDirectoryPage struct {
	Channels    []*types.NewsletterMetadata
	NextCursor  string
	HasNextPage bool
}

type directoryPageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

// GetNewsletterDirectory calls WhatsApp's channel directory query (queryNewslettersDirectory),
// which is what the app's own "Discover channels" screen uses to browse/recommend channels.
//
// This wraps an internal, reverse-engineered WhatsApp Web query that isn't exposed as a public
// whatsmeow method upstream. It may change or stop working without notice. There is no known
// free-text search parameter: this only browses/paginates recommended channels, optionally
// filtered by country.
func (cli *Client) GetNewsletterDirectory(ctx context.Context, input NewsletterDirectoryInput) (*NewsletterDirectoryPage, error) {
	data, err := cli.sendMexIQ(ctx, queryNewslettersDirectory, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}
	var resp struct {
		List struct {
			PageInfo directoryPageInfo           `json:"page_info"`
			Result   []*types.NewsletterMetadata `json:"result"`
		} `json:"xwa2_newsletters_directory_list"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &NewsletterDirectoryPage{
		Channels:    resp.List.Result,
		NextCursor:  resp.List.PageInfo.EndCursor,
		HasNextPage: resp.List.PageInfo.HasNextPage,
	}, nil
}

// GetRecommendedNewsletters calls WhatsApp's recommended-channels query (queryRecommendedNewsletters).
//
// Same caveats as GetNewsletterDirectory: internal/reverse-engineered, no free-text search.
func (cli *Client) GetRecommendedNewsletters(ctx context.Context, countryCodes []string, limit int) (*NewsletterDirectoryPage, error) {
	data, err := cli.sendMexIQ(ctx, queryRecommendedNewsletters, map[string]any{
		"input": map[string]any{"limit": limit, "country_codes": countryCodes},
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Recommended struct {
			PageInfo directoryPageInfo           `json:"page_info"`
			Result   []*types.NewsletterMetadata `json:"result"`
		} `json:"xwa2_newsletters_recommended"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &NewsletterDirectoryPage{
		Channels:    resp.Recommended.Result,
		NextCursor:  resp.Recommended.PageInfo.EndCursor,
		HasNextPage: resp.Recommended.PageInfo.HasNextPage,
	}, nil
}
