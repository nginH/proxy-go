all: build

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

run:
	@go run cmd/api/main.go

test:
	@echo "Testing..."
	@go test ./tests -v

clean:
	@echo "Cleaning..."
	@rm -f main

watch:
	@echo "seting up live reload..."
	@export PATH=$PATH:$(go env GOPATH)/bin
	@echo "path set..."
	@if command -v air > /dev/null; then \
	    air; \
	    echo "Watching...";\
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        go install github.com/cosmtrek/air@latest; \
	        air; \
	        echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

docker:
	docker build -t go-api .

docker-run:
	docker run --env-file .env -p 3000:3000 go-api


.PHONY: all build run test clean
