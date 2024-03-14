FROM golang:1.21-alpine AS base

WORKDIR /app

COPY . .
RUN go mod vendor
RUN go build main.go

FROM golang:1.21-alpine
WORKDIR /app
COPY --from=base /app/data-access/postgres_service/config.json ./data-access/postgres_service/config.json
COPY --from=base /app/main ./main

EXPOSE 47395

CMD [ "./main", "-prod" ]