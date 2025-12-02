package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Cursor represents a pagination cursor
// Supports both ID-based (uint) and UUID-based (string) cursors
type Cursor struct {
	ID        uint      `json:"id,omitempty"`         // For ID-based pagination
	CreatedAt time.Time `json:"created_at,omitempty"` // For time-based pagination
	After     string    `json:"after,omitempty"`      // For UUID or string-based pagination
	HasMore   bool      `json:"has_more"`             // Indicates if there are more results
}

// EncodeCursor encodes a cursor to a base64 string
func EncodeCursor(cursor *Cursor) (string, error) {
	if cursor == nil {
		return "", nil
	}

	data, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// DecodeCursor decodes a base64 string to a cursor
func DecodeCursor(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, err
	}

	return &cursor, nil
}

// Request represents a pagination request
type Request struct {
	Cursor    string
	Limit     int
	Direction string // "next" or "prev"
}

// ParseRequest parses pagination parameters from query strings
func ParseRequest(cursor string, limit string, direction string, defaultLimit, maxLimit int) (*Request, error) {
	// Parse limit
	limitInt := defaultLimit
	if limit != "" {
		parsedLimit, err := strconv.Atoi(limit)
		if err != nil {
			return nil, fmt.Errorf("invalid limit parameter")
		}
		limitInt = parsedLimit
	}

	// Validate limit
	if limitInt <= 0 {
		limitInt = defaultLimit
	}
	if limitInt > maxLimit {
		limitInt = maxLimit
	}

	// Validate direction
	if direction == "" {
		direction = "next"
	}
	if direction != "next" && direction != "prev" {
		return nil, fmt.Errorf("invalid direction parameter")
	}

	return &Request{
		Cursor:    cursor,
		Limit:     limitInt,
		Direction: direction,
	}, nil
}

// Response represents a pagination response
type Response struct {
	Items      interface{} `json:"items"`
	NextCursor string      `json:"next_cursor,omitempty"`
	PrevCursor string      `json:"prev_cursor,omitempty"`
	HasMore    bool        `json:"has_more"`
}

// BuildResponse builds a pagination response
func BuildResponse(items interface{}, nextCursor, prevCursor string, hasMore bool) *Response {
	return &Response{
		Items:      items,
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasMore:    hasMore,
	}
}

// OffsetPagination represents offset-based pagination
type OffsetPagination struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Limit   int   `json:"limit"` // Alias for PerPage, used in some queries
	Total   int64 `json:"total"`
}

// GetOffset calculates the offset for the current page
func (p *OffsetPagination) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

// GetLastPage calculates the last page number
func (p *OffsetPagination) GetLastPage() int {
	if p.Total == 0 {
		return 1
	}
	lastPage := int(p.Total) / p.PerPage
	if int(p.Total)%p.PerPage != 0 {
		lastPage++
	}
	return lastPage
}

// ParseOffsetRequest parses offset-based pagination parameters
func ParseOffsetRequest(page, perPage string, defaultPerPage, maxPerPage int) (*OffsetPagination, error) {
	// Parse page
	pageInt := 1
	if page != "" {
		parsedPage, err := strconv.Atoi(page)
		if err != nil {
			return nil, fmt.Errorf("invalid page parameter")
		}
		if parsedPage > 0 {
			pageInt = parsedPage
		}
	}

	// Parse per_page
	perPageInt := defaultPerPage
	if perPage != "" {
		parsedPerPage, err := strconv.Atoi(perPage)
		if err != nil {
			return nil, fmt.Errorf("invalid per_page parameter")
		}
		perPageInt = parsedPerPage
	}

	// Validate per_page
	if perPageInt <= 0 {
		perPageInt = defaultPerPage
	}
	if perPageInt > maxPerPage {
		perPageInt = maxPerPage
	}

	return &OffsetPagination{
		Page:    pageInt,
		PerPage: perPageInt,
		Limit:   perPageInt, // Set Limit as well
	}, nil
}

// Params represents generic pagination parameters
type Params struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// Result represents a generic paginated result
type Result[T any] struct {
	Data  []T   `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// GetTotalPages calculates total pages
func (r *Result[T]) GetTotalPages() int {
	if r.Total == 0 || r.Limit == 0 {
		return 0
	}
	pages := int(r.Total) / r.Limit
	if int(r.Total)%r.Limit != 0 {
		pages++
	}
	return pages
}

// HasNextPage checks if there's a next page
func (r *Result[T]) HasNextPage() bool {
	return r.Page < r.GetTotalPages()
}

// HasPrevPage checks if there's a previous page
func (r *Result[T]) HasPrevPage() bool {
	return r.Page > 1
}
