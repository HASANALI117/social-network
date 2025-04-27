# Database migrations
MIGRATE_CMD = /home/ahmed/go/bin/migrate

migrate-create:
	@$(MIGRATE_CMD) create -ext sql -dir pkg/db/migrations/sqlite -seq $(name)

migrate-up:
	@$(MIGRATE_CMD) -database "sqlite3://pkg/db/social_network.db" -path pkg/db/migrations/sqlite up

migrate-down:
@$(MIGRATE_CMD) -database "sqlite3://pkg/db/social_network.db" -path pkg/db/migrations/sqlite down

migrate-force:
	@echo "Forcing migration version to $(version)..."
	@$(MIGRATE_CMD) -database "sqlite3://pkg/db/social_network.db" -path pkg/db/migrations/sqlite force $(version)

mc: migrate-create
mu: migrate-up
md: migrate-down

# Generate Swagger documentation
swagger:
	@swag init -g cmd/server/main.go

sw: swagger
