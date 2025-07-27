.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	go build -o bin/schoolgpt ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t schoolgpt .

# Run Docker container
docker-run:
	docker run -p 8080:8080 --env-file .env schoolgpt

# Run Docker Compose
docker-compose-up:
	docker-compose up -d

# Stop Docker Compose
docker-compose-down:
	docker-compose down

# Create .env file from example
env:
	cp env.example .env 