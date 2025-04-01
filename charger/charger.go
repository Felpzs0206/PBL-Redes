package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

type Message struct {
	Action  string                 `json:"action"`
	Content map[string]interface{} `json:"content"`
}

// Definição manual do ID e porta do ponto de recarga
const (
	ID   = "ponto_1" // Pode ser alterado para ponto_2, ponto_3, etc.
	Port = ":6001"   // Porta específica para esse ponto de recarga
)

// Variáveis globais para armazenar a posição gerada
var latitude, longitude float64

func main() {
	// Gera a posição aleatória apenas uma vez na inicialização
	latitude, longitude = gerarPosicaoAleatoria()

	// Inicializa o listener na porta especificada
	listener, err := net.Listen("tcp", Port)
	if err != nil {
		fmt.Println("Erro ao iniciar o ponto de recarga:", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Ponto de Recarga %s aguardando requisições na porta %s...\n", ID, Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}
		go handleServerRequest(conn)
	}
}

func handleServerRequest(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	message, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Erro ao ler mensagem:", err)
		return
	}

	if strings.TrimSpace(message) == "LISTAR_PONTOS" {
		fmt.Printf("Servidor solicitou informações do %s.\n", ID)

		// Criando resposta em JSON com a posição fixa gerada na inicialização
		responseData := Message{
			Action: "INFORMACOES_DO_PONTO",
			Content: map[string]interface{}{
				"ID":        ID,
				"latitude":  latitude,
				"longitude": longitude,
			},
		}

		// Convertendo a estrutura para JSON
		jsonResponse, err := json.Marshal(responseData)
		if err != nil {
			fmt.Println("Erro ao converter para JSON:", err)
			return
		}

		// Enviando resposta ao servidor
		fmt.Fprintln(conn, string(jsonResponse))
		fmt.Println("Dados enviados ao Servidor:", string(jsonResponse))
	}
}

// Função para gerar latitude e longitude aleatórias uma única vez
func gerarPosicaoAleatoria() (float64, float64) {
	rand.Seed(time.Now().UnixNano())
	lat := rand.Float64()*180 - 90  // Gera um valor entre -90 e 90
	lon := rand.Float64()*360 - 180 // Gera um valor entre -180 e 180
	return lat, lon
}
