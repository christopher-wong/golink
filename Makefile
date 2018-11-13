all: build run

build:
	docker build . -t golink
.PHONY: build

run:
	docker run --rm --name golink \
		golink
.PHONY: run