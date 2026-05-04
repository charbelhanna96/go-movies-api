package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/charbelhanna96/go-movies-api/internal/handler"
	"github.com/charbelhanna96/go-movies-api/internal/model"
	"github.com/charbelhanna96/go-movies-api/internal/repository"
	commonkafka "github.com/charbelhanna96/go-movies-common/pkg/kafka"
	"github.com/charbelhanna96/go-movies-common/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMovieRepository struct {
	movies []model.Movie
	err    error
}

func (m *mockMovieRepository) GetMovies(ctx context.Context, filters repository.MovieFilters) ([]model.Movie, error) {
	return m.movies, m.err
}

type mockKafkaProducer struct{}

func (m *mockKafkaProducer) PublishSearchEvent(event commonkafka.SearchEvent) error {
	return nil
}

var testMovies = []model.Movie{
	{
		ID:           1,
		Title:        "Movie A",
		YearReleased: 2000,
		Rating:       4.50,
		DurationMins: 100,
		Genre:        model.Genre{ID: 1, Title: "Genre A"},
		Director:     model.Director{ID: 1, FirstName: "John", LastName: "Doe"},
	},
	{
		ID:           2,
		Title:        "Movie B",
		YearReleased: 2010,
		Rating:       4.80,
		DurationMins: 120,
		Genre:        model.Genre{ID: 2, Title: "Genre B"},
		Director:     model.Director{ID: 2, FirstName: "Jane", LastName: "Doe"},
	},
}

func newHandler(repo *mockMovieRepository) *handler.MoviesHandler {
	return handler.NewMoviesHandler(repo, &mockKafkaProducer{}, 5*time.Second)
}

func TestGetMovies_Success(t *testing.T) {
	repo := &mockMovieRepository{movies: testMovies}
	h := newHandler(repo)

	req := testutil.NewRequest("GET", "/api/v1/movies", nil)
	w := testutil.NewRecorder()

	h.GetMovies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []model.Movie
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Len(t, result, 2)
	assert.Equal(t, "Movie A", result[0].Title)
}

func TestGetMovies_EmptyResult(t *testing.T) {
	repo := &mockMovieRepository{movies: []model.Movie{}}
	h := newHandler(repo)

	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"min-rating": "5.0"})
	w := testutil.NewRecorder()

	h.GetMovies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []model.Movie
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Len(t, result, 0)
}

func TestGetMovies_InvalidQueryParam(t *testing.T) {
	repo := &mockMovieRepository{movies: testMovies}
	h := newHandler(repo)

	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"min-year": "abc"})
	w := testutil.NewRecorder()

	h.GetMovies(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "invalid query parameters", result["message"])
}

func TestGetMovies_RepositoryError(t *testing.T) {
	repo := &mockMovieRepository{err: fmt.Errorf("database connection lost")}
	h := newHandler(repo)

	req := testutil.NewRequest("GET", "/api/v1/movies", nil)
	w := testutil.NewRecorder()

	h.GetMovies(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var result map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "internal server error", result["message"])
}

func TestGetMovies_ContentTypeIsJSON(t *testing.T) {
	repo := &mockMovieRepository{movies: testMovies}
	h := newHandler(repo)

	req := testutil.NewRequest("GET", "/api/v1/movies", nil)
	w := testutil.NewRecorder()

	h.GetMovies(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestGetMovies_WithValidFilters(t *testing.T) {
	repo := &mockMovieRepository{movies: testMovies[:1]}
	h := newHandler(repo)

	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{
		"directors":  "1",
		"min-rating": "4.0",
		"limit":      "1",
	})
	w := testutil.NewRecorder()

	h.GetMovies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []model.Movie
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Len(t, result, 1)
}

func TestGetMovies_ResponseIsJSONArray(t *testing.T) {
	repo := &mockMovieRepository{movies: []model.Movie{}}
	h := newHandler(repo)

	req := testutil.NewRequest("GET", "/api/v1/movies", nil)
	w := testutil.NewRecorder()

	h.GetMovies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "[]")
}

func TestGetMovies_PaginationHasMore(t *testing.T) {
	repo := &mockMovieRepository{movies: testMovies}
	h := newHandler(repo)

	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"limit": "1"})
	w := testutil.NewRecorder()

	h.GetMovies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "true", w.Header().Get("X-Has-More"))
	assert.NotEmpty(t, w.Header().Get("X-Next-Cursor"))
}

func TestGetMovies_PaginationNoMore(t *testing.T) {
	repo := &mockMovieRepository{movies: testMovies}
	h := newHandler(repo)

	req := testutil.NewRequest("GET", "/api/v1/movies", map[string]string{"limit": "10"})
	w := testutil.NewRecorder()

	h.GetMovies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "false", w.Header().Get("X-Has-More"))
	assert.Empty(t, w.Header().Get("X-Next-Cursor"))
}
