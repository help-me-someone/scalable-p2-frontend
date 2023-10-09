build:
	go generate && go build -o bin/app

run: build
	./bin/app
