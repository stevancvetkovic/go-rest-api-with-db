# syntax=docker/dockerfile:1

FROM golang:1.23.3 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /rest-api-with-db


FROM scratch

COPY --from=builder /rest-api-with-db /app/rest-api-with-db

EXPOSE 8080

CMD ["/app/rest-api-with-db"]
