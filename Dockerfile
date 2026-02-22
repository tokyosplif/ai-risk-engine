FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/risk-engine ./cmd/server/main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

RUN adduser -D appuser
RUN touch prompts.json && chown appuser:appuser prompts.json

USER appuser

COPY --from=builder /bin/risk-engine .
COPY prompts.json .

EXPOSE 50051

CMD ["./risk-engine"]