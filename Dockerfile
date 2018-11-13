FROM golang:1.11.1
ADD . /go/src/golink

WORKDIR /go/src/golink

RUN go get golang.org/x/net/html/...

RUN go build .

COPY golink .