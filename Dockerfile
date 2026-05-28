# ===== BUILD STAGE =====
FROM golang:1.23-alpine AS builder

WORKDIR /app

# dependências básicas (evita erro de build em alpine)
RUN apk add --no-cache git

# copia tudo
COPY . .

# força módulos Go (evita erro go.mod ausente em builds limpos)
RUN go mod init pastebin 2>/dev/null || true
RUN go mod tidy

# build estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pastebin .

# ===== RUNTIME STAGE =====
FROM alpine:latest

WORKDIR /app

# timezone + certs (boa prática pra HTTP/TLS etc)
RUN apk add --no-cache ca-certificates

# binário
COPY --from=builder /app/pastebin .

# assets obrigatórios
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# pasta de dados
RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./pastebin"]
