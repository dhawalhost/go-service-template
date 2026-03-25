module github.com/dhawalhost/go-service-template

go 1.25

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/jackc/pgx/v5 v5.7.4
	go.uber.org/zap v1.27.0
	gorm.io/gorm v1.26.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

// TODO: remove replace directive once gokit is published to pkg.go.dev
replace github.com/dhawalhost/gokit => ../gokit
