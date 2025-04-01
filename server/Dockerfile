FROM golang:1.20 AS builder
WORKDIR /app

# Copia os arquivos do projeto
COPY go.mod .
RUN go mod tidy   
# Garante que todas as dependências sejam instaladas
COPY . .

# Compila apenas o binário do servidor
RUN go build -o server server.go

# Imagem final
FROM golang:1.20
WORKDIR /app
COPY --from=builder /app/server .

# Garante que o binário tenha permissão de execução
RUN chmod +x /app/server

CMD ["/app/server"]
