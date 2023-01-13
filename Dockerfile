FROM golang:1.18 AS build

WORKDIR /app

COPY go.mod /app/
COPY go.sum /app/

RUN go mod download

COPY . /app/

RUN GOOS=linux GO111MODULE=on CGO_ENABLED=0 go build -o me-snort3-parser ./cmd/

FROM scratch

COPY --from=build /app/me-snort3-parser /app/me-snort3-parser

CMD [ "/app/me-snort3-parser", "-b" ]
