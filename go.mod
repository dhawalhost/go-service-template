module github.com/dhawalhost/go-service-template

go 1.25

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/jackc/pgx/v5 v5.7.4
	go.uber.org/zap v1.27.0
	gorm.io/gorm v1.26.1
)

// TODO: remove replace directive once gokit is published to pkg.go.dev
replace github.com/dhawalhost/gokit => ../gokit
