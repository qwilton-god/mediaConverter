.PHONY: migrate-up migrate-down

migrate-up:
	docker run --rm -v $(PWD)/api/database/migrations:/migrations --network mediaconverter_default migrate/migrate -path /migrations -database "postgres://user:password@postgres:5432/mediadb?sslmode=disable" up

migrate-down:
	docker run --rm -v $(PWD)/api/database/migrations:/migrations --network mediaconverter_default migrate/migrate -path /migrations -database "postgres://user:password@postgres:5432/mediadb?sslmode=disable" down
