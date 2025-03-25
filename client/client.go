package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

type Message struct {
	Action  string                 `json:"action"`
	Content map[string]interface{} `json:"content"`
}

type Carro struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

const (
	listarPontosAction = "LISTAR_PONTOS"
)

func main() {
	serverAddr := "server:5000"

	// Conecta ao Servidor
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	// Envia a requisição para listar os pontos de recarga
	// fmt.Fprintln(conn, "LISTAR_PONTOS")
	carro := Carro{
		Latitude:  -23.5505,
		Longitude: -46.6333,
	}
	listarPontosMessage := listarPontos(carro)
	listarPontosJSON, _ := json.Marshal(listarPontosMessage)

	// Garante que a mensagem JSON seja enviada corretamente com um \n
	listarPontosJSON = append(listarPontosJSON, '\n')
	fmt.Printf("%#v", listarPontosJSON)

	_, err = conn.Write(listarPontosJSON)
	if err != nil {
		fmt.Println("Erro ao enviar mensagem ao servidor:", err)
		return
	}
	// Lê todas as linhas da resposta do Servidor
	fmt.Println("Posições dos Pontos de Recarga recebidas:")
	reader := bufio.NewReader(conn)
	for {
		response, err := reader.ReadString('\n')
		if err != nil {
			break // Sai do loop quando não houver mais dados
		}
		fmt.Println(strings.TrimSpace(response))
	}
}

func listarPontos(carro Carro) Message {
	return Message{
		Action: listarPontosAction,
		Content: map[string]interface{}{
			"longitude": carro.Longitude,
			"latitude":  carro.Latitude,
		},
	}
}
