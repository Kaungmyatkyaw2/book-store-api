include .envrc 




.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]



.PHONY: run/api
run/api: 
	go run ./cmd/api -db-dsn=${DB_DSN} -jwt-secret={JWT_SECRET}



.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${DB_DSN} up

.PHONY: swagger/init 
swagger/init:
	cd ./cmd/api/ && swag init