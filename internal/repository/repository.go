// Package repository defines data access contracts and implementations.
package repository

import (
	"context"

	"github.com/charbelhanna96/go-movies-api/internal/model"
)

type MovieFilters struct {
	DirectorIDs []int
	GenreIDs    []int
	MinYear     *int
	MaxYear     *int
	MinDuration *int
	MaxDuration *int
	MinRating   *float64
	MaxRating   *float64
	Limit       *int
}

type MovieRepository interface {
	GetMovies(ctx context.Context, filters MovieFilters) ([]model.Movie, error)
}
