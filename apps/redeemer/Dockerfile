FROM golang:1.16 AS build

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
COPY *.go ./
RUN go mod download

RUN go build -o redeemer
RUN go test

FROM almalinux:latest

RUN groupadd notroot
RUN useradd notroot -g notroot

WORKDIR /app
COPY --from=build /app/redeemer /app/redeemer
USER notroot:notroot

EXPOSE 8080

CMD [ "/app/redeemer" ]
