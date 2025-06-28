.PHONY: generate deldb freshdb

generate:
	@go generate ./...

deldb:
	@rm sqlite.db || true

freshdb: deldb
	@sqlite3 sqlite.db < proxy/migrations/0001_Create_schema.sql
