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

run-off-chain:
#                1         2     	   3  					4  							 5 						       
#	go run . chain_mode	network       host                 user                         pass 					
	go run . off_chain testnet3 localhost:18332 DeW+bgKg011pJHZnaBvgv/lMRks= wD9aohGo2f5LwVg7fdj1ntHQcfY= 
 
run-sim-net:
#                1         2     	   3  		 4  		   5 						 
#	go run . chain_mode	network       host      user          pass 					
	go run . off_chain simnet localhost:18554 youruser SomeDecentp4ssw0rd 

run-on-chain:
#                1         2     	   3  		 4  		   5 						 
#	go run . chain_mode	network       host      user          pass 					
	go run . on_chain testnet3 localhost:18332 DeW+bgKg011pJHZnaBvgv/lMRks= wD9aohGo2f5LwVg7fdj1ntHQcfY=

.PHONY: postgres createdb dropdb migrate_up migrate_down sqlc run-test-net run-sim-net run-on-chain

# Docker not run in case: sudo chmod 777 /var/run/docker.sock
# Permanent: https://stackoverflow.com/questions/48957195/how-to-fix-docker-got-permission-denied-issue