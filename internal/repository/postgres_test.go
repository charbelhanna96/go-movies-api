//go:build integration

package repository_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/charbelhanna96/go-movies-api/internal/repository"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()

	migrateSQL, err := filepath.Abs("../../db-migrations/migrate.sql")
	require.NoError(t, err)

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("movie"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password123"),
		postgres.WithInitScripts(migrateSQL),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	require.NoError(t, db.Ping())

	return db
}

func TestGetMovies_NoFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
}

func TestGetMovies_OrderedByRatingDesc(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{})

	require.NoError(t, err)
	require.True(t, len(movies) > 1)

	for i := 1; i < len(movies); i++ {
		assert.GreaterOrEqual(t, movies[i-1].Rating, movies[i].Rating,
			"movies should be ordered by rating descending")
	}
}

func TestGetMovies_FilterByDirector(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	directorID := 1
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		DirectorIDs: []int{directorID},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
	for _, m := range movies {
		assert.Equal(t, directorID, m.Director.ID)
	}
}

func TestGetMovies_FilterByGenre(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	genreID := 1
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		GenreIDs: []int{genreID},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
	for _, m := range movies {
		assert.Equal(t, genreID, m.Genre.ID)
	}
}

func TestGetMovies_FilterByMinYear(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	minYear := 2010
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		MinYear: &minYear,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
	for _, m := range movies {
		assert.GreaterOrEqual(t, m.YearReleased, minYear)
	}
}

func TestGetMovies_FilterByMaxYear(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	maxYear := 2000
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		MaxYear: &maxYear,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
	for _, m := range movies {
		assert.LessOrEqual(t, m.YearReleased, maxYear)
	}
}

func TestGetMovies_FilterByYearRange(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	minYear := 2000
	maxYear := 2015
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		MinYear: &minYear,
		MaxYear: &maxYear,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
	for _, m := range movies {
		assert.GreaterOrEqual(t, m.YearReleased, minYear)
		assert.LessOrEqual(t, m.YearReleased, maxYear)
	}
}

func TestGetMovies_FilterByMinDuration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	minDuration := 150
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		MinDuration: &minDuration,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
	for _, m := range movies {
		assert.GreaterOrEqual(t, m.DurationMins, minDuration)
	}
}

func TestGetMovies_FilterByMaxDuration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	maxDuration := 120
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		MaxDuration: &maxDuration,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
	for _, m := range movies {
		assert.LessOrEqual(t, m.DurationMins, maxDuration)
	}
}

func TestGetMovies_FilterByMinRating(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	minRating := 4.5
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		MinRating: &minRating,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
	for _, m := range movies {
		assert.GreaterOrEqual(t, m.Rating, minRating)
	}
}

func TestGetMovies_FilterByMaxRating(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	maxRating := 4.3
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		MaxRating: &maxRating,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
	for _, m := range movies {
		assert.LessOrEqual(t, m.Rating, maxRating)
	}
}

func TestGetMovies_FilterByLimit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	limit := 3
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		Limit: &limit,
	})

	require.NoError(t, err)
	assert.Len(t, movies, limit)
}

func TestGetMovies_MultipleDirectors(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		DirectorIDs: []int{1, 2},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
	for _, m := range movies {
		assert.Contains(t, []int{1, 2}, m.Director.ID)
	}
}

func TestGetMovies_MultipleGenres(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		GenreIDs: []int{1, 5},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, movies)
	for _, m := range movies {
		assert.Contains(t, []int{1, 5}, m.Genre.ID)
	}
}

func TestGetMovies_NoMatchingFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	minYear := 2100
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		MinYear: &minYear,
	})

	require.NoError(t, err)
	assert.Empty(t, movies)
}

func TestGetMovies_CombinedFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	minYear := 2000
	minRating := 4.5
	limit := 5
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		DirectorIDs: []int{1},
		MinYear:     &minYear,
		MinRating:   &minRating,
		Limit:       &limit,
	})

	require.NoError(t, err)
	assert.LessOrEqual(t, len(movies), limit)
	for _, m := range movies {
		assert.Equal(t, 1, m.Director.ID)
		assert.GreaterOrEqual(t, m.YearReleased, minYear)
		assert.GreaterOrEqual(t, m.Rating, minRating)
	}
}

func TestGetMovies_ReturnsEmptySliceNotNil(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewPostgresMovieRepository(db)
	minYear := 2100
	movies, err := repo.GetMovies(context.Background(), repository.MovieFilters{
		MinYear: &minYear,
	})

	require.NoError(t, err)
	assert.NotNil(t, movies)
	assert.Empty(t, movies)
}
