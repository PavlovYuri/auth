run:
	go run ./cmd/auth --config=./config/local_config.yaml

run-test:
	go run ./cmd/auth --config=./config/local_tests_config.yaml

migrate:
	go run ./cmd/migrator --storage-path=./storage/auth.db --migrations-path=./migrations

migrate-test:
	go run ./cmd/migrator --storage-path=./storage/auth.db --migrations-path=./tests/migrations --migrations-table=migrations_test

test:
	go test ./tests -count=1 -v