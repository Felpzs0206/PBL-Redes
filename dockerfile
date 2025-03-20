FROM golang:1.20
WORKDIR /app
COPY . .

# Compila cada programa separadamente
RUN go build -o server server.go
RUN go build -o client client.go
RUN go build -o charger charger.go
