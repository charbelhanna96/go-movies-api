package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/charbelhanna96/go-movies-api/internal/repository"
	"github.com/charbelhanna96/go-movies-api/internal/validator"
	"github.com/charbelhanna96/go-movies-api/internal/web"
)

func NewMoviesHandler(movieRepo repository.MovieRepository, timeout time.Duration) *MoviesHandler {
	return &MoviesHandler{
		movieRepo: movieRepo,
		timeout:   timeout,
	}
}

type MoviesHandler struct {
	movieRepo repository.MovieRepository
	timeout   time.Duration
}

// GetMovies handles GET /api/v1/movies requests.
// It returns a JSON array of matching movies or an error if the request fails.
func (handler *MoviesHandler) GetMovies(respWriter http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), handler.timeout)
	defer cancel()

	filters, err := validator.ParseMovieFilters(req)
	if err != nil {
		web.Error(respWriter, http.StatusBadRequest, "invalid query parameters")
		return
	}

	slog.Debug("movie filters parsed", "filters", filters)

	movies, err := handler.movieRepo.GetMovies(ctx, filters)
	if err != nil {
		slog.Error("failed to get movies", "error", err, "path", req.URL.Path, "query", req.URL.RawQuery)
		web.Error(respWriter, http.StatusInternalServerError, "internal server error")
		return
	}

	web.JSON(respWriter, http.StatusOK, movies)
}
