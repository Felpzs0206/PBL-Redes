# Usa a imagem oficial do Golang como base
FROM golang:1.20 AS builder
WORKDIR /app

# Copia os arquivos do projeto
COPY go.mod .
RUN go mod tidy   

# Copia todo o código-fonte para o container
COPY . .

# Compila o binário do charger
RUN go build -o charger charger.go

# Imagem final mais leve
FROM golang:1.20
WORKDIR /app

# Copia apenas o binário compilado
COPY --from=builder /app/charger .

# Garante permissão de execução
RUN chmod +x /app/charger

# Executa o charger
CMD ["/app/charger"]
