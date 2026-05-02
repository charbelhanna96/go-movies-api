package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/charbelhanna96/go-movies-api/internal/metrics"
	"github.com/charbelhanna96/go-movies-api/internal/model"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type postgresMovieRepository struct {
	db *sql.DB
}

// NewPostgresMovieRepository creates a repository backed by PostgreSQL.
func NewPostgresMovieRepository(db *sql.DB) MovieRepository {
	return &postgresMovieRepository{db: db}
}

// GetMovies retrieves movies from the database based on the provided filters.
func (r *postgresMovieRepository) GetMovies(ctx context.Context, filters MovieFilters) ([]model.Movie, error) {
	tracer := otel.Tracer("go-movies-api")
	ctx, span := tracer.Start(ctx, "db.QueryMovies")
	defer span.End()

	start := time.Now()

	query, args := buildQuery(filters)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query movies: %w", err)
	}
	defer rows.Close()

	movies := make([]model.Movie, 0)
	for rows.Next() {
		var movie model.Movie
		if err := rows.Scan(
			&movie.ID, &movie.Title, &movie.YearReleased, &movie.Rating, &movie.DurationMins,
			&movie.Genre.ID, &movie.Genre.Title,
			&movie.Director.ID, &movie.Director.FirstName, &movie.Director.LastName,
		); err != nil {
			return nil, fmt.Errorf("scan movie: %w", err)
		}
		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate movies: %w", err)
	}

	span.SetAttributes(attribute.Int("movies.count", len(movies)))

	metrics.DatabaseQueryDuration.WithLabelValues("get_movies").Observe(time.Since(start).Seconds())

	return movies, nil
}

// buildQuery creates a parameterized SQL query from movie filters.
func buildQuery(filters MovieFilters) (string, []any) {
	// INNER JOIN is safe here as all movies have valid genre and director references.
	// In production with nullable foreign keys, LEFT JOIN would be safer.
	query := `
	SELECT
		m.id,
		m.title,
		m.year_released,
		m.rating,
		m.duration_mins,
		g.id,
		g.title,
		d.id,
		d.first_name,
		d.last_name
	FROM movie m
	JOIN genre g ON g.id = m.genre_id
	JOIN director d ON d.id = m.director_id
	`

	var conditions []string
	var args []any
	argIndex := 1

	if len(filters.DirectorIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("m.director_id = ANY($%d)", argIndex))
		args = append(args, pq.Array(filters.DirectorIDs))
		argIndex++
	}

	if len(filters.GenreIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("m.genre_id = ANY($%d)", argIndex))
		args = append(args, pq.Array(filters.GenreIDs))
		argIndex++
	}

	if filters.MinYear != nil {
		conditions = append(conditions, fmt.Sprintf("m.year_released >= $%d", argIndex))
		args = append(args, *filters.MinYear)
		argIndex++
	}

	if filters.MaxYear != nil {
		conditions = append(conditions, fmt.Sprintf("m.year_released <= $%d", argIndex))
		args = append(args, *filters.MaxYear)
		argIndex++
	}

	if filters.MinDuration != nil {
		conditions = append(conditions, fmt.Sprintf("m.duration_mins >= $%d", argIndex))
		args = append(args, *filters.MinDuration)
		argIndex++
	}

	if filters.MaxDuration != nil {
		conditions = append(conditions, fmt.Sprintf("m.duration_mins <= $%d", argIndex))
		args = append(args, *filters.MaxDuration)
		argIndex++
	}

	if filters.MinRating != nil {
		conditions = append(conditions, fmt.Sprintf("m.rating >= $%d", argIndex))
		args = append(args, *filters.MinRating)
		argIndex++
	}

	if filters.MaxRating != nil {
		conditions = append(conditions, fmt.Sprintf("m.rating <= $%d", argIndex))
		args = append(args, *filters.MaxRating)
		argIndex++
	}

	if filters.AfterRating != nil && filters.AfterID != nil {
		conditions = append(conditions, fmt.Sprintf(
			"(m.rating < $%d OR (m.rating = $%d AND m.id > $%d))",
			argIndex, argIndex+1, argIndex+2,
		))
		args = append(args, *filters.AfterRating, *filters.AfterRating, *filters.AfterID)
		argIndex += 3
	}

	if len(conditions) > 0 {
		query += "\nWHERE " + strings.Join(conditions, " AND ")
	}

	query += "\nORDER BY m.rating DESC, m.id ASC"

	if filters.Limit != nil {
		query += fmt.Sprintf("\nLIMIT $%d", argIndex)
		args = append(args, *filters.Limit)
	}

	return query, args
}
