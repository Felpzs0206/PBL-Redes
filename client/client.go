package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Action  string                 `json:"action"`
	Content map[string]interface{} `json:"content"`
}

type Carro struct {
	ID        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Bateria   int     `json:"bateria"`
}

const (
	listarPontosAction       = "LISTAR_PONTOS"
	fazerReservaAction       = "RESERVAR_PONTO"
	inicioCarregamentoAction = "INICIO_CARREGAMENTO"
	fimCarregamentoAction    = "FIM_CARREGAMENTO"
)

var (
	serverAddr    = "server:5000"
	mutex         sync.Mutex // Proteção contra condições de corrida
	commandChan   = make(chan string)
	alertaEnviado bool

	carro = Carro{
		ID:        "carro-" + os.Getenv("HOSTNAME") + "-" + strconv.Itoa(rand.Intn(1000)),
		Latitude:  rand.Float64()*180 - 90,
		Longitude: rand.Float64()*360 - 180,
		Bateria:   100,
	}
)

func main() {
	go monitorarBateria(commandChan) // Bateria agora usa o mesmo canal
	go entradaUsuario(commandChan)   // Entrada do usuário

	for cmd := range commandChan {
		switch cmd {
		case "1":
			fmt.Println("\nListando pontos de recarga...")
			enviarMensagem(listarPontos(carro))
		case "2":
			fmt.Println("\nReservando ponto de recarga...")
			enviarMensagem(reservarPonto(carro))
		case "3":
			fmt.Println("\nInformando início do carregamento...")
			enviarMensagem(inicioCarregamento(carro))
		case "4":
			fmt.Println("\nInformando fim do carregamento...")
			enviarMensagem(fimCarregamento(carro))
		case "BATERIA_CRITICA":
			fmt.Println("\nBateria em nível crítico! Conecte-se a um ponto de recarga.")
			enviarMensagem(listarPontos(carro))
		default:
			fmt.Println("\nOpção inválida. Escolha uma opção válida.")
		}
		mostrarMenu()
	}
}

// Mostra o menu de opções para o usuário
func mostrarMenu() {
	fmt.Println("\n==== MENU PRINCIPAL ====")
	fmt.Println("1. Listar pontos de recarga")
	fmt.Println("2. Reservar ponto de recarga")
	fmt.Println("3. Informar início do carregamento")
	fmt.Println("4. Informar fim do carregamento")
	fmt.Print("Escolha uma opção: ")
}

// Monitora a bateria e reduz ao longo do tempo
func monitorarBateria(commandChan chan<- string) {
	for {
		time.Sleep(5 * time.Second) // consumo de bateria
		mutex.Lock()
		carro.Bateria -= 10
		if carro.Bateria < 0 {
			carro.Bateria = 0
		}
		fmt.Printf("\nBateria - Nível atual: %d%%\n", carro.Bateria)

		// envia alerta apenas uma vez quando a bateria chega a 20%
		if carro.Bateria <= 20 && !alertaEnviado {
			commandChan <- "BATERIA_CRITICA"
			alertaEnviado = true
		}
		mutex.Unlock()
	}
}

// Captura entrada do usuário
func entradaUsuario(commandChan chan<- string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao ler entrada:", err)
			continue
		}
		commandChan <- strings.TrimSpace(input)
	}
}

// 🚀 Envia mensagens para o servidor
func enviarMensagem(msg Message) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	jsonMsg, _ := json.Marshal(msg)
	jsonMsg = append(jsonMsg, '\n')

	_, err = conn.Write(jsonMsg)
	if err != nil {
		fmt.Println("Erro ao enviar mensagem ao servidor:", err)
		return
	}

	fmt.Println("Mensagem enviada ao servidor:", string(jsonMsg))

	// Lê a resposta do servidor
	reader := bufio.NewReader(conn)
	for {
		response, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println("Resposta do servidor:", strings.TrimSpace(response))
	}
}

// Cria uma mensagem JSON para listar pontos de recarga
func listarPontos(carro Carro) Message {
	return Message{
		Action: listarPontosAction,
		Content: map[string]interface{}{
			"ID":        carro.ID,
			"longitude": carro.Longitude,
			"latitude":  carro.Latitude,
		},
	}
}

// Cria uma mensagem JSON para reservar um ponto
func reservarPonto(carro Carro) Message {
	return Message{
		Action: fazerReservaAction,
		Content: map[string]interface{}{
			"ID": carro.ID,
		},
	}
}

// Informa o início do carregamento
func inicioCarregamento(carro Carro) Message {
	return Message{
		Action: inicioCarregamentoAction,
		Content: map[string]interface{}{
			"ID": carro.ID,
		},
	}
}

// Informa o fim do carregamento
func fimCarregamento(carro Carro) Message {
	return Message{
		Action: fimCarregamentoAction,
		Content: map[string]interface{}{
			"ID": carro.ID,
		},
	}
}
