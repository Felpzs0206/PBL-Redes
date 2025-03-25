FROM golang:1.20 AS builder
WORKDIR /app

# Copia os arquivos do projeto
COPY go.mod .
RUN go mod tidy  # Garante que todas as dependências sejam instaladas

COPY . .

# Compila os binários do servidor e do charger
RUN go build -o server server.go
RUN go build -o charger charger.go

# Imagem final
FROM golang:1.20
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/charger .

# Garante que os binários tenham permissão de execução
RUN chmod +x /app/server /app/charger

CMD ["/app/charger"]

