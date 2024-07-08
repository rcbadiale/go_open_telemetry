# Go Expert - Open Telemetry

## Run locally

Setup your your [Weather API](https://www.weatherapi.com/) key on `docker-compose.yml`.

In the project root execute:
```shell
docker compose up
```

Make a request:
```shell
curl -X POST -H "Content-Type: application/json" -d '{"cep": "13405162"}' http://localhost:8080/weather
```

Go to http://localhost:9411/
Click on `Run Query`

## APIs

### POST /weather

Examples are available at `api/requests.http`

200:
```json
Request Body:
{"cep":"12345678"}

Response Body:
{"city":"somewhere","temp_C":16,"temp_F":60.8,"temp_K":289.1}
```

422:
```
invalid zipcode
```

## URLs

| Service    | URL                    |
| ---------- | ---------------------- |
| Jaeger     | http://localhost:16686 |
| Zipkin     | http://localhost:9411  |
| Webserver  | http://localhost:8080  |
