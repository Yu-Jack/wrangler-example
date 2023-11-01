install:
	kubectl apply -f config/crd

resource:
	kubectl apply -f resources

run:
	go run cmd/main.go

all: install resource run
