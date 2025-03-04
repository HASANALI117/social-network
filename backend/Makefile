# Database migrations
mc:
	@migrate create -ext sql -dir pkg/db/migrations/sqlite -seq $(name)

mu:
	@migrate -database "sqlite3://pkg/db/social_network.db" -path pkg/db/migrations/sqlite up

md:
	@migrate -database "sqlite3://pkg/db/social_network.db" -path pkg/db/migrations/sqlite down

migrate-create: mc
migrate-up: mu
migrate-down: md