FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dineware-engine .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/dineware-engine ./dineware-engine
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./dineware-engine"]
