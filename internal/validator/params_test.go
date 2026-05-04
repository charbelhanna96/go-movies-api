package validator_test

import (
	"strings"
	"testing"

	"github.com/charbelhanna96/go-movies-api/internal/validator"
	"github.com/charbelhanna96/go-movies-common/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMovieFilters_NoParams(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", nil)
	filters, err := validator.ParseMovieFilters(req)
	require.NoError(t, err)
	assert.Nil(t, filters.DirectorIDs)
	assert.Nil(t, filters.GenreIDs)
	assert.Nil(t, filters.MinYear)
	assert.Nil(t, filters.MaxYear)
	assert.Nil(t, filters.MinDuration)
	assert.Nil(t, filters.MaxDuration)
	assert.Nil(t, filters.MinRating)
	assert.Nil(t, filters.MaxRating)
	assert.Nil(t, filters.Limit)
}

func TestParseMovieFilters_ValidDirectors(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"directors": "1,2,3"})
	filters, err := validator.ParseMovieFilters(req)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, filters.DirectorIDs)
}

func TestParseMovieFilters_ValidGenres(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"genres": "1,5"})
	filters, err := validator.ParseMovieFilters(req)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 5}, filters.GenreIDs)
}

func TestParseMovieFilters_ValidYearRange(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{
		"min-year": "2000",
		"max-year": "2020",
	})
	filters, err := validator.ParseMovieFilters(req)
	require.NoError(t, err)
	assert.Equal(t, 2000, *filters.MinYear)
	assert.Equal(t, 2020, *filters.MaxYear)
}

func TestParseMovieFilters_ValidDurationRange(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{
		"min-duration": "90",
		"max-duration": "180",
	})
	filters, err := validator.ParseMovieFilters(req)
	require.NoError(t, err)
	assert.Equal(t, 90, *filters.MinDuration)
	assert.Equal(t, 180, *filters.MaxDuration)
}

func TestParseMovieFilters_ValidRatingRange(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{
		"min-rating": "4.0",
		"max-rating": "5.0",
	})
	filters, err := validator.ParseMovieFilters(req)
	require.NoError(t, err)
	assert.Equal(t, 4.0, *filters.MinRating)
	assert.Equal(t, 5.0, *filters.MaxRating)
}

func TestParseMovieFilters_ValidLimit(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"limit": "10"})
	filters, err := validator.ParseMovieFilters(req)
	require.NoError(t, err)
	assert.Equal(t, 10, *filters.Limit)
}

func TestParseMovieFilters_DeduplicatesDirectorIDs(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"directors": "1,2,1,3"})
	filters, err := validator.ParseMovieFilters(req)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, filters.DirectorIDs)
}

func TestParseMovieFilters_InvalidDirectorID(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"directors": "abc"})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_InvalidMinYear(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"min-year": "abc"})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_InvalidMinRating(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"min-rating": "abc"})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_MinYearExceedsMaxYear(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{
		"min-year": "2020",
		"max-year": "2000",
	})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_MinDurationExceedsMaxDuration(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{
		"min-duration": "180",
		"max-duration": "90",
	})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_MinRatingExceedsMaxRating(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{
		"min-rating": "5.0",
		"max-rating": "1.0",
	})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_YearBelowMinimum(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"min-year": "1800"})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_RatingAboveMaximum(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"min-rating": "6.0"})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_EmptyDirectorValue(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"directors": ""})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_TooManyIDs(t *testing.T) {
	ids := make([]string, 101)
	for i := range ids {
		ids[i] = "1"
	}
	raw := strings.Join(ids, ",")
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"directors": raw})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_ZeroDirectorID(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"directors": "0"})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_NegativeDirectorID(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"directors": "-1"})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_LimitAboveMaximum(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"limit": "10001"})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_DurationAboveMaximum(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"min-duration": "601"})
	_, err := validator.ParseMovieFilters(req)
	assert.Error(t, err)
}

func TestParseMovieFilters_AllFilters(t *testing.T) {
	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{
		"directors":    "1,2",
		"genres":       "3,4",
		"min-year":     "2000",
		"max-year":     "2020",
		"min-duration": "90",
		"max-duration": "180",
		"min-rating":   "4.0",
		"max-rating":   "5.0",
		"limit":        "10",
	})
	filters, err := validator.ParseMovieFilters(req)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, filters.DirectorIDs)
	assert.Equal(t, []int{3, 4}, filters.GenreIDs)
	assert.Equal(t, 2000, *filters.MinYear)
	assert.Equal(t, 2020, *filters.MaxYear)
	assert.Equal(t, 90, *filters.MinDuration)
	assert.Equal(t, 180, *filters.MaxDuration)
	assert.Equal(t, 4.0, *filters.MinRating)
	assert.Equal(t, 5.0, *filters.MaxRating)
	assert.Equal(t, 10, *filters.Limit)
}
