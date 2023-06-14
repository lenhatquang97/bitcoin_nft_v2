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

run-test-net:
#                1         2     	   3  					4  							 5 						       6
#	go run . chain_mode	network       host                 user                         pass 					send_address
	go run . off_chain testnet3 localhost:18332 DeW+bgKg011pJHZnaBvgv/lMRks= wD9aohGo2f5LwVg7fdj1ntHQcfY= n1Nd8J38uyDRLwh5ShAAPvbNrqBD1wee8v

run-sim-net:
#                1         2     	   3  		 4  		   5 						 6
#	go run . chain_mode	network       host      user          pass 					send_address
	go run . off_chain simnet localhost:18554 youruser SomeDecentp4ssw0rd SfF7WYPTkHnjx1jKweNYAoFGnhZH1Q2291

run-on-chain:
#                1         2     	   3  		 4  		   5 						 6
#	go run . chain_mode	network       host      user          pass 					send_address
	go run . on_chain testnet3 localhost:18332 DeW+bgKg011pJHZnaBvgv/lMRks= wD9aohGo2f5LwVg7fdj1ntHQcfY= n1Nd8J38uyDRLwh5ShAAPvbNrqBD1wee8v

.PHONY: postgres createdb dropdb migrate_up migrate_down sqlc run-test-net run-sim-net run-on-chain

# Docker not run in case: sudo chmod 777 /var/run/docker.sock