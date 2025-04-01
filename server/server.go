package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
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
}

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
		go handleClient(conn) // Permite múltiplos clientes simultâneos
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Lê a requisição do Cliente
	message, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Erro ao ler mensagem do cliente:", err)
		return
	}

	var request Message
	err = json.Unmarshal([]byte(message), &request)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON. A mensagem recebida foi:", message)
		return
	}

	switch request.Action {
	case "LISTAR_PONTOS":
		fmt.Println("Cliente solicitou a lista de pontos de recarga.")

		// Extrai a posição do carro
		carro := request.Content
		fmt.Println("Dados do carro recebidos:")
		fmt.Printf("ID: %v\n", carro["ID"])
		fmt.Printf("Latitude: %v\n", carro["latitude"])
		fmt.Printf("Longitude: %v\n", carro["longitude"])

		pontosRecarga := obterPontoDeRecarga("charger:6001")
		fmt.Printf("Ponto de recarga encontrado: %s\n", pontosRecarga.ID)
		fmt.Printf("Latitude: %f\n", pontosRecarga.Latitude)
		fmt.Printf("Longitude: %f\n", pontosRecarga.Longitude)
		fmt.Printf("Distância: %f km\n", calcularDistancia(carro["latitude"].(float64), carro["longitude"].(float64), pontosRecarga.Latitude, pontosRecarga.Longitude))

	case "RESERVAR_PONTO":
		fmt.Println("Cliente solicitou reserva de ponto de recarga.")

		carro := request.Content
		fmt.Println("Dados do carro recebidos:")
		fmt.Printf("ID: %v\n", carro["ID"])

	case "INICIO_CARREGAMENTO":
		fmt.Println("Cliente solicitou início de carregamento.")
		carro := request.Content
		fmt.Println("Dados do carro recebidos:")
		fmt.Printf("ID: %v\n", carro["ID"])

	case "FIM_CARREGAMENTO":
		fmt.Println("Cliente solicitou fim de carregamento.")
		carro := request.Content
		fmt.Println("Dados do carro recebidos:")
		fmt.Printf("ID: %v\n", carro["ID"])
	default:
		fmt.Println("Ação desconhecida:", request.Action)
		return
	}
	// TODO
	// para o ponto
	// solicitar disponibilidade dos pontos
	// acrescentar carro na fila
	// atualizar fila (após carregamento)

	//para o carro
	// lista de pontos ordenada
	// confirmar reserva
	// informar valor
	// confirmar pagamento

}

func obterPontoDeRecarga(endereco string) PontoRecarga {
	conn, err := net.Dial("tcp", endereco)
	if err != nil {
		fmt.Printf("Erro ao conectar ao ponto de recarga %s: %v\n", endereco, err)
		return PontoRecarga{}
	}
	defer conn.Close()

	fmt.Fprintln(conn, "LISTAR_PONTOS")

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Printf("Erro ao receber resposta do ponto %s: %v\n", endereco, err)
		return PontoRecarga{}
	}

	// Divide a resposta recebida: "ponto_1 -23.5505 -46.6333"
	parts := strings.Fields(response)
	fmt.Printf("%#v", parts)
	if len(parts) != 3 {
		fmt.Println("Formato de resposta inválido:", response)
		return PontoRecarga{}
	}

	// Converte os valores para float
	lat, err1 := strconv.ParseFloat(parts[1], 64)
	lon, err2 := strconv.ParseFloat(parts[2], 64)
	if err1 != nil || err2 != nil {
		fmt.Println("Erro ao converter coordenadas:", response)
		return PontoRecarga{}
	}

	// Retorna os dados corretamente formatados
	return PontoRecarga{
		ID:        parts[0],
		Latitude:  lat,
		Longitude: lon,
		Distancia: 0,
	}
}

func calcularDistancia(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Raio da Terra em km
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c // Retorna a distância em km
}

func calcularValorConta(tempoDecorrido float64) float64 {
	return tempoDecorrido * 0.5
}
