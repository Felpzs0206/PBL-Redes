package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
)

type Message struct {
	Action  string                 `json:"action"`
	Content map[string]interface{} `json:"content"`
}

// Lista de Pontos de Recarga e suas portas
var pontosDeRecarga = []string{"charger:6001"}

type PontoRecarga struct {
	ID        string
	Latitude  float64
	Longitude float64
	Distancia float64
	Fila      []string // Adicionado campo para a fila de espera
}

var (
	carrosEmCarregamento = make(map[string]string) // Mapa carroID -> pontoID
	carregamentoMutex    sync.Mutex
)

func main() {
	listener, err := net.Listen("tcp", ":5000")
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Servidor ouvindo na porta 5000...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	message, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Erro ao ler mensagem do cliente:", err)
		return
	}

	var request Message
	err = json.Unmarshal([]byte(message), &request)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	switch request.Action {
	case "LISTAR_PONTOS":
		handleListarPontos(conn, request)
	case "RESERVAR_PONTO":
		handleReservarPonto(conn, request)
	case "INICIO_CARREGAMENTO":
		handleInicioCarregamento(conn, request)
	case "FIM_CARREGAMENTO":
		handleFimCarregamento(conn, request)
	default:
		fmt.Println("Ação desconhecida:", request.Action)
		sendErrorResponse(conn, "Ação desconhecida")
	}
}

func handleListarPontos(conn net.Conn, request Message) {
	fmt.Println("Cliente solicitou a lista de pontos de recarga.")
	carro := request.Content

	// Obter informações de todos os pontos de recarga
	var pontos []PontoRecarga
	for _, endereco := range pontosDeRecarga {
		ponto := obterInformacoesPonto(endereco)
		if ponto.ID != "" { // Verifica se obteve resposta válida
			ponto.Distancia = calcularDistancia(
				carro["latitude"].(float64),
				carro["longitude"].(float64),
				ponto.Latitude,
				ponto.Longitude,
			)
			pontos = append(pontos, ponto)
		}
	}

	// Ordenar pontos por distância (mais próximo primeiro)
	for i := 0; i < len(pontos); i++ {
		for j := i + 1; j < len(pontos); j++ {
			if pontos[i].Distancia > pontos[j].Distancia {
				pontos[i], pontos[j] = pontos[j], pontos[i]
			}
		}
	}

	// Preparar resposta
	response := Message{
		Action: "LISTA_PONTOS",
		Content: map[string]interface{}{
			"pontos": pontos,
		},
	}

	sendResponse(conn, response)
}

func handleReservarPonto(conn net.Conn, request Message) {
	fmt.Println("Cliente solicitou reserva de ponto de recarga.")
	carro := request.Content
	carroID := carro["ID"].(string)
	pontoID := carro["pontoID"].(string)

	// Encontrar o endereço do ponto desejado
	var enderecoPonto string
	for _, endereco := range pontosDeRecarga {
		if strings.Contains(endereco, pontoID) {
			enderecoPonto = endereco
			break
		}
	}

	if enderecoPonto == "" {
		sendErrorResponse(conn, "Ponto de recarga não encontrado")
		return
	}

	// Enviar comando RESERVAR_PONTO para o ponto específico
	connPonto, err := net.Dial("tcp", enderecoPonto)
	if err != nil {
		sendErrorResponse(conn, fmt.Sprintf("Erro ao conectar ao ponto: %v", err))
		return
	}
	defer connPonto.Close()

	// Enviar comando de reserva
	msgReserva := Message{
		Action: "RESERVAR_PONTO",
		Content: map[string]interface{}{
			"carroID": carroID,
		},
	}

	sendResponse(connPonto, msgReserva)
	// Ler resposta do ponto
	responsePonto, err := bufio.NewReader(connPonto).ReadString('\n')
	if err != nil {
		sendErrorResponse(conn, fmt.Sprintf("Erro ao ler resposta do ponto: %v", err))
		return
	}

	// Decodificar resposta do ponto
	var respostaPonto Message
	if err := json.Unmarshal([]byte(responsePonto), &respostaPonto); err != nil {
		sendErrorResponse(conn, "Resposta do ponto inválida")
		return
	}

	// Encaminhar resposta ao cliente
	sendResponse(conn, respostaPonto)
}

func handleInicioCarregamento(conn net.Conn, request Message) {
	fmt.Println("Cliente solicitou início de carregamento.")
	carro := request.Content
	carroID := carro["ID"].(string)
	pontoID := carro["pontoID"].(string)

	carregamentoMutex.Lock()
	defer carregamentoMutex.Unlock()

	// Verificar se o ponto está disponível
	if _, exists := carrosEmCarregamento[pontoID]; exists {
		sendErrorResponse(conn, "Ponto já está em uso")
		return
	}

	// Registrar carregamento
	carrosEmCarregamento[pontoID] = carroID

	response := Message{
		Action: "CARREGAMENTO_INICIADO",
		Content: map[string]interface{}{
			"pontoID": pontoID,
			"carroID": carroID,
		},
	}

	fmt.Println("Carregamento iniciado:", pontoID, carroID)
	fmt.Println(response)
	sendResponse(conn, response)
}

func handleFimCarregamento(conn net.Conn, request Message) {
	fmt.Println("Cliente solicitou fim de carregamento.")
	carro := request.Content
	carroID := carro["ID"].(string)
	pontoID := carro["pontoID"].(string)
	tempo := carro["tempo"].(float64)

	carregamentoMutex.Lock()
	defer carregamentoMutex.Unlock()

	// Verificar se o carregamento existe
	if currentCarID, exists := carrosEmCarregamento[pontoID]; !exists || currentCarID != carroID {
		sendErrorResponse(conn, "Carregamento não encontrado")
		return
	}

	// Calcular valor
	valor := calcularValorConta(tempo)

	// Remover do registro
	delete(carrosEmCarregamento, pontoID)

	response := Message{
		Action: "CARREGAMENTO_FINALIZADO",
		Content: map[string]interface{}{
			"valor": valor,
		},
	}

	fmt.Println(response)

	sendResponse(conn, response)
}

func obterInformacoesPonto(endereco string) PontoRecarga {
	conn, err := net.Dial("tcp", endereco)
	if err != nil {
		fmt.Printf("Erro ao conectar ao ponto %s: %v\n", endereco, err)
		return PontoRecarga{}
	}
	defer conn.Close()

	// Enviar comando LISTAR_PONTOS
	fmt.Fprintln(conn, "LISTAR_PONTOS")

	// Ler resposta
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Printf("Erro ao ler resposta do ponto %s: %v\n", endereco, err)
		return PontoRecarga{}
	}

	// Decodificar resposta JSON
	var msg Message
	if err := json.Unmarshal([]byte(response), &msg); err != nil {
		fmt.Printf("Resposta inválida do ponto %s: %v\n", endereco, err)
		return PontoRecarga{}
	}

	if msg.Action != "INFORMACOES_DO_PONTO" {
		fmt.Printf("Resposta inesperada do ponto %s: %s\n", endereco, msg.Action)
		return PontoRecarga{}
	}

	content := msg.Content
	return PontoRecarga{
		ID:        content["ID"].(string),
		Latitude:  content["latitude"].(float64),
		Longitude: content["longitude"].(float64),
		Fila:      convertInterfaceToStringSlice(content["fila"]),
	}
}

func convertInterfaceToStringSlice(data interface{}) []string {
	if data == nil {
		return nil
	}
	slice := data.([]interface{})
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = v.(string)
	}
	return result
}

func calcularDistancia(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Raio da Terra em km
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func calcularValorConta(tempoDecorrido float64) float64 {
	return tempoDecorrido * 0.5
}

func sendResponse(conn net.Conn, response Message) {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Erro ao codificar resposta:", err)
		return
	}
	fmt.Fprintln(conn, string(jsonResponse))
}

func sendErrorResponse(conn net.Conn, message string) {
	response := Message{
		Action: "ERRO",
		Content: map[string]interface{}{
			"mensagem": message,
		},
	}
	sendResponse(conn, response)
}
