FROM golang:1.24-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM base AS development
COPY . .
EXPOSE 8080
CMD ["go", "run", "./cmd/server"]

FROM base AS build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.21 AS production
WORKDIR /app
COPY --from=build /server ./server
EXPOSE 8080
CMD ["./server"]
