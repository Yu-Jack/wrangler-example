install:
	kubectl apply -f resources

run:
	go run cmd/main.go

all: install run