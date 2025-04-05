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

// ERROS
// Cliente consegue se colocar na fila de espera mais de uma vez
// Cliente ainda não escolhe qual ponto deseja reservar

type Message struct {
	Action  string                 `json:"action"`
	Content map[string]interface{} `json:"content"`
}

type Carro struct {
	ID        string      `json:"id"`
	Porta     string      `json:"porta"`
	Latitude  float64     `json:"latitude"`
	Longitude float64     `json:"longitude"`
	Bateria   int         `json:"bateria"`
	Historico []Historico `json:"historico"`
}

type Historico struct {
	ID                   string               `json:"id"`
	SessaoDeCarregamento SessaoDeCarregamento `json:"sessao_de_carregamento"`
	Pagamento            Pagamento            `json:"pagamento"`
}

type SessaoDeCarregamento struct {
	Inicio time.Time `json:"inicio"`
	Fim    time.Time `json:"fim"`
}

type Pagamento struct {
	Valor float64 `json:"valor"`
	Pago  bool    `json:"status"`
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
	porta         = os.Getenv("PORTA")

	carro = Carro{
		ID:        "carro-" + os.Getenv("HOSTNAME") + "-" + strconv.Itoa(rand.Intn(1000)),
		Porta:     porta,
		Latitude:  rand.Float64()*180 - 90,
		Longitude: rand.Float64()*360 - 180,
		Bateria:   100,
		Historico: []Historico{},
	}
)

func main() {
	// go monitorarBateria(commandChan) // Bateria agora usa o mesmo canal
	go entradaUsuario(commandChan) // Entrada do usuário
	mostrarMenu()
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
			enviarMensagem(inicioCarregamento(&carro))
		case "4":
			fmt.Println("\nInformando fim do carregamento...")
			enviarMensagem(fimCarregamento(&carro))
		case "BATERIA_CRITICA":
			fmt.Println("\nBateria em nível crítico! Conecte-se a um ponto de recarga.")
			enviarMensagem(listarPontos(carro))
		default:
			fmt.Println("\nOpção inválida. Escolha uma opção válida.")
		}
		mostrarMenu()
	}
}

// receber a response e tratar para lidar com os actions
func handleServerResponse(r string) {
	//TODO
	// Formatar os prints
	// Tratar CARREGAMENTO_INICIADO e RESERVA_CONFIRMADA

	var response Message
	err := json.Unmarshal([]byte(r), &response)
	if err != nil {
		fmt.Println("Erro ao decodificar a resposta do servidor:", err)
		return
	}
	switch response.Action {
	case "CARREGAMENTO_FINALIZADO":
		fmt.Println("Valor do pagamento recebido:", response.Content["valor"])
		carro.adicionarPagamento(response.Content["valor"].(float64))
		fmt.Println("Pagamento adicionado ao histórico do carro.")
		fmt.Println(carro)
	case "LISTA_PONTOS":
		//TODO
		// Cliente deve escolher qual ponto deseja reservar
		fmt.Println(response)
		pontos := response.Content["pontos"].([]interface{})
		fmt.Println("Lista de pontos de recarga disponíveis:")
		for _, ponto := range pontos {
			pontoMap := ponto.(map[string]interface{})
			fmt.Printf("ID: %s, Distância: %.2f km\n", pontoMap["ID"], pontoMap["Distancia"])
		}
	default:
		fmt.Println("Ação não reconhecida:", response.Action)
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
		if carro.Bateria < 10 {
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

// Envia mensagens para o servidor
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
		handleServerResponse(response)
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
// Usuário deverá informar qual ponto deseja reservar
func reservarPonto(carro Carro) Message {
	return Message{
		Action: fazerReservaAction,
		Content: map[string]interface{}{
			"ID":      carro.ID,
			"pontoID": "charger:6001",
		},
	}
}

// Informa o início do carregamento
func inicioCarregamento(c *Carro) Message {
	novoHistorico := Historico{
		ID: fmt.Sprintf("sessao-%d", time.Now().Unix()),
		SessaoDeCarregamento: SessaoDeCarregamento{
			Inicio: time.Now(),
		},
	}

	fmt.Println(novoHistorico.SessaoDeCarregamento.Inicio)

	c.Historico = append(c.Historico, novoHistorico)

	return Message{
		Action: "INICIO_CARREGAMENTO",
		Content: map[string]interface{}{
			"ID":      c.ID,
			"pontoID": "charger:6001",
		},
	}
}

// 	// Adiciona ao histórico do carro
// 	c.Historico = append(c.Historico, novaSessao)

// 	fmt.Println("Sessão de carregamento iniciada.")

// 	return Message{
// 		Action: inicioCarregamentoAction,
// 		Content: map[string]interface{}{
// 			"ID": c.ID,
// 		},
// 	}
// }

// Informa o fim do carregamento
func fimCarregamento(c *Carro) Message {
	c.Historico[len(c.Historico)-1].SessaoDeCarregamento.Fim = time.Now()

	tempoDecorrido := calcularTempoDecorrido(c.Historico[len(c.Historico)-1].SessaoDeCarregamento.Inicio,
		c.Historico[len(c.Historico)-1].SessaoDeCarregamento.Fim)

	return Message{
		Action: fimCarregamentoAction,
		Content: map[string]interface{}{
			"ID":      c.ID,
			"pontoID": "charger:6001",
			"tempo":   tempoDecorrido,
		},
	}
}

// }
func calcularTempoDecorrido(inicio, fim time.Time) float64 {
	return fim.Sub(inicio).Seconds()
}

func (c *Carro) adicionarPagamento(valor float64) {
	c.Historico[len(c.Historico)-1].Pagamento.Valor = valor
	c.Historico[len(c.Historico)-1].Pagamento.Pago = false
	fmt.Println("Pagamento adicionado ao histórico do carro.")
}
