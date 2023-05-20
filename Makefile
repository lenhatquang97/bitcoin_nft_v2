postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root nft_collection

dropdb:
	docker exec -it postgres12 dropdb nft_collection

migrate_up:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/nft_collection?sslmode=disable" -verbose up

migrate_down:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/nft_collection?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

.PHONY: postgres createdb dropdb migrate_up migrate_down sqlc