package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/charbelhanna96/go-movies-api/internal/kafka"
	"github.com/charbelhanna96/go-movies-api/internal/metrics"
	"github.com/charbelhanna96/go-movies-api/internal/repository"
	"github.com/charbelhanna96/go-movies-api/internal/validator"
	"github.com/charbelhanna96/go-movies-api/internal/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func NewMoviesHandler(movieRepo repository.MovieRepository, kafkaProducer kafka.KafkaProducer, timeout time.Duration) *MoviesHandler {
	return &MoviesHandler{
		movieRepo:     movieRepo,
		timeout:       timeout,
		kafkaProducer: kafkaProducer,
	}
}

type MoviesHandler struct {
	movieRepo     repository.MovieRepository
	timeout       time.Duration
	kafkaProducer kafka.KafkaProducer
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

	// fetch one extra to detect if there are more results
	limitPlusOne := filters.Limit
	if limitPlusOne != nil {
		n := *limitPlusOne + 1
		limitPlusOne = &n
	}
	filtersWithExtra := filters
	filtersWithExtra.Limit = limitPlusOne

	ctx, repoSpan := tracer.Start(ctx, "GetMovies")
	repoSpan.SetAttributes(
		attribute.Int("directors.count", len(filters.DirectorIDs)),
		attribute.Int("genres.count", len(filters.GenreIDs)),
	)

	movies, err := handler.movieRepo.GetMovies(ctx, filtersWithExtra)
	if err != nil {
		repoSpan.SetStatus(codes.Error, err.Error())
		repoSpan.End()
		slog.Error("failed to get movies", "error", err, "path", req.URL.Path, "query", req.URL.RawQuery)
		web.Error(respWriter, http.StatusInternalServerError, "internal server error")
		return
	}
	repoSpan.End()

	// check if there are more results
	hasMore := filters.Limit != nil && len(movies) > *filters.Limit
	if hasMore {
		movies = movies[:*filters.Limit]
	}

	// set pagination headers
	respWriter.Header().Set("X-Has-More", strconv.FormatBool(hasMore))
	if hasMore && len(movies) > 0 {
		last := movies[len(movies)-1]
		respWriter.Header().Set("X-Next-Cursor", encodeCursor(last.Rating, last.ID))
	}
	// publish search event to Kafka — non-blocking, errors are logged not returned
	go func() {
		event := kafka.SearchEvent{
			Filters: kafka.SearchFilters{
				DirectorIDs: filters.DirectorIDs,
				GenreIDs:    filters.GenreIDs,
				MinYear:     filters.MinYear,
				MaxYear:     filters.MaxYear,
				MinDuration: filters.MinDuration,
				MaxDuration: filters.MaxDuration,
				MinRating:   filters.MinRating,
				MaxRating:   filters.MaxRating,
			},
			ResultsCount: len(movies),
			Timestamp:    time.Now().UTC(),
		}
		if err := handler.kafkaProducer.PublishSearchEvent(event); err != nil {
			slog.Error("failed to publish search event", "error", err)
		}
	}()
	metrics.MoviesReturned.Observe(float64(len(movies)))
	web.JSON(respWriter, http.StatusOK, movies)
}

// encodeCursor encodes the last movie's rating and id into a base64 cursor string.
func encodeCursor(rating float64, id int) string {
	raw := fmt.Sprintf("%g:%d", rating, id)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}
