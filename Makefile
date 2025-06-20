.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


.PHONY: run/api
run/api: 
	go run ./cmd/api


.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}


.PHONY: db/psql
db/psql:
	psql ${DB_DSN}

.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${DB_DSN} up



.PHONY: docker/build
docker/build: confirm 
	docker-compose --env-file .env.docker up --build

.PHONY: swagger/generate 
swagger/generate:
	cd ./cmd/api/ && swag init

