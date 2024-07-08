## Build stage
FROM golang:alpine AS builder

WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o input_service cmd/open_telemetry/input_service/main.go

## Run stage
FROM alpine

# Settings
RUN apk add -U tzdata ca-certificates
ENV TZ=America/Sao_Paulo
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime

COPY --from=builder /app/input_service /input_service

ENTRYPOINT [ "/input_service" ]
