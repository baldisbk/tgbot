coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o cover.html

compose:
	docker-compose build
	docker-compose up