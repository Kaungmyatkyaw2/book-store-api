include .envrc 

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


.PHONY: run/api
run/api: 
	go run ./cmd/api -db-dsn=${DB_DSN} -jwt-secret=${JWT_SECRET} -oauth-redirect-url=${GOOGLE_OAUTH_REDIRECT_URL} -oauth-client-id=${GOOGLE_OAUTH_CLIENT_ID} -oauth-client-secret=${GOOGLE_OAUTH_CLIENT_SECRET}


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

.PHONY: swagger/generate 
swagger/generate:
	cd ./cmd/api/ && swag init