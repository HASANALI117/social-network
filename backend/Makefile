# Database migrations
migrate-create:
	@migrate create -ext sql -dir pkg/db/migrations/sqlite -seq $(name)

migrate-up:
	@migrate -database "sqlite3://pkg/db/social_network.db" -path pkg/db/migrations/sqlite up

migrate-down:
	@migrate -database "sqlite3://pkg/db/social_network.db" -path pkg/db/migrations/sqlite down

mc: migrate-create
mu: migrate-up
md: migrate-down

# Generate Swagger documentation
swagger:
	@swag init -g cmd/server/main.go

sw: swagger