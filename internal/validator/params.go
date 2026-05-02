// Package validator provides functions for parsing and validating query parameters for the API endpoints.
package validator

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/charbelhanna96/go-movies-api/internal/repository"
	playgroundValidator "github.com/go-playground/validator/v10"
)

// maxIDsPerFilter caps the number of IDs accepted per list filter.
const maxIDsPerFilter = 100

var validate = playgroundValidator.New()

// movieFilterParams defines the expected query parameters for filtering movies.
type movieFilterParams struct {
	DirectorIDs []int    `validate:"dive,gt=0"`
	GenreIDs    []int    `validate:"dive,gt=0"`
	MinYear     *int     `validate:"omitempty,min=1888,max=2100"`
	MaxYear     *int     `validate:"omitempty,min=1888,max=2100"`
	MinDuration *int     `validate:"omitempty,min=1,max=600"`
	MaxDuration *int     `validate:"omitempty,min=1,max=600"`
	MinRating   *float64 `validate:"omitempty,min=0,max=5"`
	MaxRating   *float64 `validate:"omitempty,min=0,max=5"`
	Limit       *int     `validate:"omitempty,min=1,max=1000"`
	AfterCursor *string  `validate:"omitempty"`
}

// ParseMovieFilters extracts and validates the query parameters from the HTTP request.
func ParseMovieFilters(r *http.Request) (repository.MovieFilters, error) {
	query := r.URL.Query()

	directorIDs, err := parseIDList(query, "directors")
	if err != nil {
		return repository.MovieFilters{}, fmt.Errorf("invalid directors: %w", err)
	}

	genreIDs, err := parseIDList(query, "genres")
	if err != nil {
		return repository.MovieFilters{}, fmt.Errorf("invalid genres: %w", err)
	}

	minYear, err := parseOptionalInt(query, "min-year")
	if err != nil {
		return repository.MovieFilters{}, fmt.Errorf("invalid min-year: %w", err)
	}

	maxYear, err := parseOptionalInt(query, "max-year")
	if err != nil {
		return repository.MovieFilters{}, fmt.Errorf("invalid max-year: %w", err)
	}

	minDuration, err := parseOptionalInt(query, "min-duration")
	if err != nil {
		return repository.MovieFilters{}, fmt.Errorf("invalid min-duration: %w", err)
	}

	maxDuration, err := parseOptionalInt(query, "max-duration")
	if err != nil {
		return repository.MovieFilters{}, fmt.Errorf("invalid max-duration: %w", err)
	}

	minRating, err := parseOptionalFloat(query, "min-rating")
	if err != nil {
		return repository.MovieFilters{}, fmt.Errorf("invalid min-rating: %w", err)
	}

	maxRating, err := parseOptionalFloat(query, "max-rating")
	if err != nil {
		return repository.MovieFilters{}, fmt.Errorf("invalid max-rating: %w", err)
	}

	limit, err := parseOptionalInt(query, "limit")
	if err != nil {
		return repository.MovieFilters{}, fmt.Errorf("invalid limit: %w", err)
	}

	params := movieFilterParams{
		DirectorIDs: directorIDs,
		GenreIDs:    genreIDs,
		MinYear:     minYear,
		MaxYear:     maxYear,
		MinDuration: minDuration,
		MaxDuration: maxDuration,
		MinRating:   minRating,
		MaxRating:   maxRating,
		Limit:       limit,
	}

	if err := validate.Struct(params); err != nil {
		return repository.MovieFilters{}, fmt.Errorf("validation failed: %w", err)
	}

	if params.MinYear != nil && params.MaxYear != nil && *params.MinYear > *params.MaxYear {
		return repository.MovieFilters{}, fmt.Errorf("min-year must not exceed max-year")
	}

	if params.MinDuration != nil && params.MaxDuration != nil && *params.MinDuration > *params.MaxDuration {
		return repository.MovieFilters{}, fmt.Errorf("min-duration must not exceed max-duration")
	}

	if params.MinRating != nil && params.MaxRating != nil && *params.MinRating > *params.MaxRating {
		return repository.MovieFilters{}, fmt.Errorf("min-rating must not exceed max-rating")
	}

	var afterRating *float64
	var afterID *int

	afterCursor := strings.TrimSpace(query.Get("after_cursor"))
	if query.Has("after_cursor") {
		if afterCursor == "" {
			return repository.MovieFilters{}, fmt.Errorf("invalid after_cursor: value is required when \"after_cursor\" is provided")
		}
		afterRating, afterID, err = parseCursor(afterCursor)
		if err != nil {
			return repository.MovieFilters{}, fmt.Errorf("invalid after_cursor: %w", err)
		}
	}

	return repository.MovieFilters{
		DirectorIDs: params.DirectorIDs,
		GenreIDs:    params.GenreIDs,
		MinYear:     params.MinYear,
		MaxYear:     params.MaxYear,
		MinDuration: params.MinDuration,
		MaxDuration: params.MaxDuration,
		MinRating:   params.MinRating,
		MaxRating:   params.MaxRating,
		Limit:       params.Limit,
		AfterRating: afterRating,
		AfterID:     afterID,
	}, nil
}

// parseIDList converts a comma-separated query parameter into a deduplicated slice of ints.
func parseIDList(query url.Values, key string) ([]int, error) {
	if !query.Has(key) {
		return nil, nil
	}

	raw := strings.TrimSpace(query.Get(key))
	if raw == "" {
		return nil, fmt.Errorf("value is required when %q is provided", key)
	}

	parts := strings.Split(raw, ",")
	if len(parts) > maxIDsPerFilter {
		return nil, fmt.Errorf("too many IDs: maximum is %d", maxIDsPerFilter)
	}

	seen := make(map[int]struct{}, len(parts))
	ids := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("empty token in list")
		}

		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("%q is not a valid integer", part)
		}

		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// parseOptionalInt converts a query parameter to an integer pointer, returning nil when the parameter is omitted.
func parseOptionalInt(query url.Values, key string) (*int, error) {
	if !query.Has(key) {
		return nil, nil
	}

	raw := strings.TrimSpace(query.Get(key))
	if raw == "" {
		return nil, fmt.Errorf("value is required when %q is provided", key)
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil, fmt.Errorf("%q is not a valid integer", raw)
	}

	return &value, nil
}

// parseOptionalFloat converts a query parameter to a float64 pointer, returning nil when the parameter is omitted.
func parseOptionalFloat(query url.Values, key string) (*float64, error) {
	if !query.Has(key) {
		return nil, nil
	}

	raw := strings.TrimSpace(query.Get(key))
	if raw == "" {
		return nil, fmt.Errorf("value is required when %q is provided", key)
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, fmt.Errorf("%q is not a valid number", raw)
	}

	return &value, nil
}

// parseCursor decodes a base64 cursor string into rating and id components.
func parseCursor(cursor string) (*float64, *int, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid cursor encoding")
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("invalid cursor format")
	}

	rating, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid cursor rating")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid cursor id")
	}

	return &rating, &id, nil
}
