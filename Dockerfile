FROM golang:1.16

WORKDIR /go/src/main

COPY . .

RUN go mod download
RUN go mod verify
RUN go build -o main

ENTRYPOINT [ "./main" ]
