package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/charbelhanna96/go-movies-api/internal/metrics"
	"github.com/charbelhanna96/go-movies-api/internal/repository"
	"github.com/charbelhanna96/go-movies-api/internal/validator"
	"github.com/charbelhanna96/go-movies-api/internal/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

	tracer := otel.Tracer("go-movies-api")

	// span for validation
	ctx, validateSpan := tracer.Start(ctx, "ParseMovieFilters")
	filters, err := validator.ParseMovieFilters(req)
	if err != nil {
		validateSpan.SetStatus(codes.Error, err.Error())
		validateSpan.End()
		web.Error(respWriter, http.StatusBadRequest, "invalid query parameters")
		return
	}
	validateSpan.End()

	slog.Debug("movie filters parsed", "filters", filters)

	// span for repository call
	ctx, repoSpan := tracer.Start(ctx, "GetMovies")
	repoSpan.SetAttributes(
		attribute.Int("directors.count", len(filters.DirectorIDs)),
		attribute.Int("genres.count", len(filters.GenreIDs)),
	)
	movies, err := handler.movieRepo.GetMovies(ctx, filters)

	if err != nil {
		slog.Error("failed to get movies", "error", err, "path", req.URL.Path, "query", req.URL.RawQuery)
		repoSpan.SetStatus(codes.Error, err.Error())
		repoSpan.End()
		web.Error(respWriter, http.StatusInternalServerError, "internal server error")
		return
	}
	repoSpan.End()

	metrics.MoviesReturned.Observe(float64(len(movies)))

	web.JSON(respWriter, http.StatusOK, movies)
}
