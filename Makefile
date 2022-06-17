bench:
	go test -cpu 1,2,4,8,10 -benchmem -run=^$ -bench . .

run:
	MODE=ec go run ./...

run_aes:
	MODE=aes go run ./...
