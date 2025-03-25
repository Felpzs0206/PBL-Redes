package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Defina manualmente a porta e a posição do Ponto de Recarga
const (
	ID        = "ponto_1" // Mude para ponto_2, ponto_3 conforme necessário
	Latitude  = -23.5505  // Latitude real do ponto de recarga
	Longitude = -46.6333  // Longitude real do ponto de recarga
	Port      = ":6001"   // Porta associada a esse ponto
)

func main() {
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
		fmt.Printf("Servidor solicitou posição do %s.\n", ID)

		// Responde com o ID, latitude e longitude como floats reais
		response := fmt.Sprintf("%s %.6f %.6f\n", ID, Latitude, Longitude)
		fmt.Fprintln(conn, response)
		fmt.Println("Dados enviados ao Servidor:", response)
	}

	// TODO
	// retorna posição e fila de espera
}
