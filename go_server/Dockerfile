FROM golang:1.16 AS build

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./

RUN go test
RUN go build -o redeemer

FROM almalinux:latest

RUN groupadd notroot
RUN useradd notroot -g notroot

WORKDIR /app
COPY --from=build /app/redeemer /app/redeemer
USER notroot:notroot

EXPOSE 8080

CMD [ "/app/redeemer" ]
