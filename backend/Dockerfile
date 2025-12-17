FROM golang:1.23-alpine AS build

WORKDIR /app

COPY backend/go.mod ./go.mod
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/karen-api ./cmd/server

FROM alpine:3.20
RUN adduser -D -H -s /sbin/nologin app
USER app

WORKDIR /app
COPY --from=build /out/karen-api ./karen-api

EXPOSE 8080
CMD ["./karen-api"]


