package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

type Message struct {
	Action  string                 `json:"action"`
	Content map[string]interface{} `json:"content"`
}

// Variáveis globais
var (
	ID                  = os.Getenv("ID")   // Pode ser alterado para ponto_2, ponto_3, etc.
	Port                = os.Getenv("PORT") // Porta específica para esse ponto de recarga
	latitude, longitude float64
	waitingQueue        []string   // Fila de espera para carros
	queueMutex          sync.Mutex // Mutex para proteger acesso concorrente à fila
)

func main() {
	// Gera a posição aleatória apenas uma vez na inicialização
	latitude, longitude = gerarPosicaoAleatoria()

	// Inicializa o listener na porta especificada
	// Garante que a porta tenha o formato ":6001"
	if Port != "" && Port[0] != ':' {
		Port = ":" + Port
	}

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
	//message = strings.TrimSpace(message)
	fmt.Println("Mensagem recebida do Servidor:", message)

	// Estrutura para armazenar o JSON recebido
	var msg Message
	err = json.Unmarshal([]byte(message), &msg)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	// Processa a ação recebida
	switch msg.Action {
	case "LISTAR_PONTOS":
		handleListarPontos(conn)
	case "RESERVAR_PONTO":
		handleReservarPonto(conn, msg.Content)

	default:
		fmt.Println("Comando não reconhecido:", msg.Action)
	}
}

func handleListarPontos(conn net.Conn) {
	fmt.Printf("Servidor solicitou informações do %s.\n", ID)

	// Criando resposta em JSON com a posição fixa gerada na inicialização
	responseData := Message{
		Action: "INFORMACOES_DO_PONTO",
		Content: map[string]interface{}{
			"ID":        ID,
			"latitude":  latitude,
			"longitude": longitude,
			"fila":      getWaitingQueue(), // Mostra o estado atual da fila
		},
	}

	sendResponse(conn, responseData)
}

func handleReservarPonto(conn net.Conn, content map[string]interface{}) {
	// Obtém o ID do carro do JSON
	carID, ok := content["carroID"].(string)
	if !ok {
		fmt.Println("Erro: campo 'carroID' ausente ou inválido")
		return
	}

	// Adiciona o carro à fila de espera
	queueMutex.Lock()
	waitingQueue = append(waitingQueue, carID)
	currentPosition := len(waitingQueue)
	queueMutex.Unlock()

	fmt.Printf("Carro %s adicionado à fila do ponto %s. Posição na fila: %d\n", carID, ID, currentPosition)

	// Resposta ao servidor
	responseData := Message{
		Action: "RESERVA_CONFIRMADA",
		Content: map[string]interface{}{
			"ID":           ID,
			"carroID":      carID,
			"posicao_fila": currentPosition,
			"latitude":     latitude,
			"longitude":    longitude,
		},
	}

	sendResponse(conn, responseData)
}

func sendResponse(conn net.Conn, responseData Message) {
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

// Função segura para obter cópia da fila de espera
func getWaitingQueue() []string {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	// Retorna uma cópia da fila para evitar acesso concorrente
	queueCopy := make([]string, len(waitingQueue))
	copy(queueCopy, waitingQueue)
	return queueCopy
}

// Função para gerar latitude e longitude aleatórias uma única vez
func gerarPosicaoAleatoria() (float64, float64) {
	rand.Seed(time.Now().UnixNano())
	lat := rand.Float64()*180 - 90  // Gera um valor entre -90 e 90
	lon := rand.Float64()*360 - 180 // Gera um valor entre -180 e 180
	return lat, lon
}
