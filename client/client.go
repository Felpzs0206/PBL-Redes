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
	ID             string      `json:"id"`
	Porta          string      `json:"porta"`
	Latitude       float64     `json:"latitude"`
	Longitude      float64     `json:"longitude"`
	Bateria        int         `json:"bateria"`
	Historico      []Historico `json:"historico"`
	EmFila         bool        `json:"em_fila"`
	PontoReservado string      `json:"ponto_reservado"`
	isCarregando   bool        `json:"is_carregando"`
}

type Historico struct {
	ID                   string               `json:"historicoID"`
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
	serverAddr             = "server:5000"
	mutex                  sync.Mutex // Proteção contra condições de corrida
	commandChan            = make(chan string)
	alertaEnviado          bool
	porta                  = os.Getenv("PORTA")
	ultimosPontosRecebidos []map[string]interface{}

	carro = Carro{
		ID:           "carro-" + os.Getenv("HOSTNAME") + "-" + strconv.Itoa(rand.Intn(1000)),
		Porta:        porta,
		Latitude:     rand.Float64()*180 - 90,
		Longitude:    rand.Float64()*360 - 180,
		Bateria:      100,
		Historico:    []Historico{},
		EmFila:       false,
		isCarregando: false,
	}
)

func main() {
	go entradaUsuario(commandChan)   // Captura entrada do usuário
	go monitorarBateria(commandChan) // Monitora a bateria

	var modoReserva bool = false // Modo de reserva de ponto
	mostrarMenu()
	for cmd := range commandChan {
		if modoReserva {
			escolha, err := strconv.Atoi(cmd)
			if err != nil || escolha < 1 || escolha > len(ultimosPontosRecebidos) {
				fmt.Println("Escolha inválida. Digite o número do ponto mostrado na lista.")
				continue
			}
			pontoEscolhido := ultimosPontosRecebidos[escolha-1]
			fmt.Printf("Reservando ponto de recarga: %s\n", pontoEscolhido["ID"])

			msg := Message{
				Action: "RESERVAR_PONTO",
				Content: map[string]interface{}{
					"ID":      carro.ID,
					"pontoID": pontoEscolhido["ID"],
					"EmFila":  carro.EmFila,
				},
			}
			enviarMensagem(msg)
			modoReserva = false
			continue
		}

		switch strings.ToUpper(cmd) {
		// case "L":
		// 	fmt.Println("\nListando pontos de recarga...")
		// 	enviarMensagem(listarPontos(carro))

		case "R":
			if len(ultimosPontosRecebidos) == 0 {
				fmt.Println("Você precisa listar os pontos antes de reservar.")
				continue
			}
			fmt.Println("Digite o número do ponto que deseja reservar:")
			modoReserva = true

		case "I":
			fmt.Println("\n-> Informando início do carregamento...")
			enviarMensagem(inicioCarregamento(&carro))

		case "F":
			fmt.Println("\n-> Informando fim do carregamento...")
			enviarMensagem(fimCarregamento(&carro))

		case "B":
			fmt.Println("\n-> Bateria em nível crítico! Listando pontos...")
			enviarMensagem(listarPontos(carro))
		case "P":
			fmt.Println("\n-> Pagando última pendência...")
			if msg := pagarUltimaPendenciaEmAberto(carro.Historico, carro.ID); msg != nil {
				enviarMensagem(*msg)
			}

		default:
			fmt.Println("\n-> Comando inválido. Use B, R, I, F ou P.")
		}
		mostrarMenu()
	}
}

// receber a response e tratar para lidar com os actions
func handleServerResponse(r string) {
	//TODO
	// Formatar os prints

	var response Message
	err := json.Unmarshal([]byte(r), &response)
	if err != nil {
		fmt.Println("Erro ao decodificar a resposta do servidor:", err)
		return
	}
	switch response.Action {
	case "CARREGAMENTO_FINALIZADO":
		handleCarregamentoFinalizado(response.Content)
	case "LISTA_PONTOS":
		ultimosPontosRecebidos = handleListaPontos(response.Content)
		fmt.Println("Pontos de recarga listados com sucesso!")
	case "RESERVA_CONFIRMADA":
		handleReservaConfirmada(response.Content)
	case "CARREGAMENTO_INICIADO":
		handleCarregamentoInciado(response.Content)
	case "PAGAMENTO_CONFIRMADO":
		handlePagamentoConfirmado(response, &carro.Historico)
	case "ERRO":
		fmt.Println("Erro:", response.Content["mensagem"])
	default:
		fmt.Println("Ação não reconhecida:", response.Action)
	}
}

func handleCarregamentoInciado(content map[string]interface{}) {
	fmt.Println("Carregamento iniciado com sucesso!")
	carro.isCarregando = true
	fmt.Println("ID do ponto de recarga:", content["pontoID"])
}

func handleCarregamentoFinalizado(content map[string]interface{}) {
	fmt.Println("Carregamento finalizado com sucesso!")
	fmt.Println("Valor do pagamento:", content["valor"])
	carro.adicionarPagamento(content["valor"].(float64))
	fmt.Println("Pagamento adicionado ao histórico do carro.")
	carro.isCarregando = false
	carro.EmFila = false
	carro.Bateria = 100
	fmt.Println(carro)
}

func handleReservaConfirmada(content map[string]interface{}) {
	fmt.Println("Reserva confirmada com sucesso!")
	carro.EmFila = true
	carro.PontoReservado = content["ID"].(string)
	fmt.Println("Ponto reservado:", carro.PontoReservado)
}

func handleListaPontos(content map[string]interface{}) []map[string]interface{} {
	pontosRaw := content["pontos"].([]interface{})
	var pontosFormatados []map[string]interface{}

	fmt.Println("\nLista de pontos de recarga disponíveis:")
	for i, ponto := range pontosRaw {
		pontoMap := ponto.(map[string]interface{})

		distancia, okDist := pontoMap["Distancia"].(float64)
		tamanhoFilaFloat, okFila := pontoMap["TamanhoFila"].(float64) // <- atenção aqui!

		if !okDist || !okFila {
			fmt.Printf("%d) Dados inválidos para o ponto\n", i+1)
			continue
		}

		fmt.Printf(
			"%d) ID: %s, Distância: %.2f km, Fila: %d carro(s)\n",
			i+1,
			pontoMap["ID"],
			distancia,
			int(tamanhoFilaFloat), // conversão segura
		)

		pontosFormatados = append(pontosFormatados, pontoMap)
	}

	fmt.Println("Digite 'R' para reservar um ponto, Em seguida, informe o do número do ponto na lista (ex: 1, 2...).")
	return pontosFormatados
}

func pagarUltimaPendenciaEmAberto(historico []Historico, carroID string) *Message {
	// Percorre o histórico de trás para frente
	for i := len(historico) - 1; i >= 0; i-- {
		if !historico[i].Pagamento.Pago {
			msg := Message{
				Action: "PAGAR_PENDENCIA",
				Content: map[string]interface{}{
					"carroID":     carroID,
					"historicoID": historico[i].ID,
					"valor":       historico[i].Pagamento.Valor,
				},
			}
			return &msg
		}
	}

	// Se não houver nenhuma pendência
	fmt.Println("Nenhuma pendência em aberto para pagamento.")
	return nil
}

func handlePagamentoConfirmado(msg Message, historico *[]Historico) {
	historicoID, ok := msg.Content["historicoID"].(string)
	if !ok {
		fmt.Println("Erro: historicoID inválido ou ausente na resposta do servidor.")
		return
	}

	for i := range *historico {
		if (*historico)[i].ID == historicoID {
			if (*historico)[i].Pagamento.Pago {
				fmt.Printf("Sessão %s já está marcada como paga.\n", historicoID)
			} else {
				(*historico)[i].Pagamento.Pago = true
				fmt.Printf("Pagamento da sessão %s confirmado e marcado como pago.\n", historicoID)
			}
			return
		}
	}

	fmt.Printf("Sessão %s não encontrada no histórico local.\n", historicoID)
}

// Mostra o menu de opções para o usuário
func mostrarMenu() {
	fmt.Println("\n--- MENU ---")
	fmt.Println("B - Bateria crítica")
	fmt.Println("R - Reservar ponto de recarga")
	fmt.Println("I - Iniciar carregamento")
	fmt.Println("F - Finalizar carregamento")
	fmt.Println("P - Pagar última pendência")
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

		if carro.Bateria == 0 {
			mutex.Unlock()
			continue
		}
		fmt.Printf("\nBateria - Nível atual: %d%%\n", carro.Bateria)

		// envia alerta apenas uma vez quando a bateria chega a 20%
		if carro.Bateria <= 20 && !alertaEnviado {
			commandChan <- "B"
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

// Informa o início do carregamento
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
			"pontoID": c.PontoReservado, // <- USANDO A RESERVA REAL
		},
	}
}

// Informa o fim do carregamento
func fimCarregamento(c *Carro) Message {
	c.Historico[len(c.Historico)-1].SessaoDeCarregamento.Fim = time.Now()

	tempoDecorrido := calcularTempoDecorrido(
		c.Historico[len(c.Historico)-1].SessaoDeCarregamento.Inicio,
		c.Historico[len(c.Historico)-1].SessaoDeCarregamento.Fim,
	)

	return Message{
		Action: fimCarregamentoAction,
		Content: map[string]interface{}{
			"ID":           c.ID,
			"pontoID":      c.PontoReservado,
			"tempo":        tempoDecorrido,
			"isCarregando": c.isCarregando,
		},
	}
}

func calcularTempoDecorrido(inicio, fim time.Time) float64 {
	return fim.Sub(inicio).Seconds()
}

func (c *Carro) adicionarPagamento(valor float64) {
	c.Historico[len(c.Historico)-1].Pagamento.Valor = valor
	c.Historico[len(c.Historico)-1].Pagamento.Pago = false
	fmt.Println("Pagamento adicionado ao histórico do carro.")
}
