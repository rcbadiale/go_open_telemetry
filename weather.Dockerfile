## Build stage
FROM golang:alpine AS builder

WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o weather_service cmd/open_telemetry/internal_service/main.go

## Run stage
FROM alpine

# Settings
RUN apk add -U tzdata ca-certificates
ENV TZ=America/Sao_Paulo
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime

COPY --from=builder /app/weather_service /weather_service

ENTRYPOINT [ "/weather_service" ]
