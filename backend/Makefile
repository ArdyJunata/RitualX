DB_URL ?= postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable

.PHONY: migrate-up migrate-down migrate-create

migrate-up:
	migrate -path migrations/ -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations/ -database "$(DB_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations/ -seq $(name)
