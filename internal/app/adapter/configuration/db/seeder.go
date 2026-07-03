package db

import "context"

type Seeder interface {
	Name() string
	Seed(ctx context.Context) error
}
