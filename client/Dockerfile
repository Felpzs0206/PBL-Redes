FROM golang:1.20 AS builder
WORKDIR /app

# Copia os arquivos do cliente
COPY go.mod .
RUN go mod tidy

COPY . .
RUN go build -o client client.go

# Imagem final
FROM golang:1.20
WORKDIR /app
COPY --from=builder /app/client .
CMD ["/app/client"]
