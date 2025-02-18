package seeder

type Seeder interface {
	Name() string
	Seed() error
}
