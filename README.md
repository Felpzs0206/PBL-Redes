# Sistema de Carregamento de Veículos Elétricos

Um sistema distribuído para gerenciamento de recarga de veículos elétricos, implementado em Go utilizando arquitetura de microserviços com comunicação via TCP.

## Visão Geral

Este projeto implementa um simulador de sistema de carregamento para veículos elétricos, composto por três tipos de componentes principais:

1. **Servidor Central**: Coordena a comunicação entre veículos e pontos de recarga
2. **Pontos de Recarga**: Gerenciam filas de espera e processos de carregamento
3. **Clientes (Veículos)**: Simulam carros elétricos que necessitam de recarga

Os componentes se comunicam via TCP/IP usando mensagens em formato JSON, permitindo a simulação de todo o ciclo de recarga de veículos elétricos.

## Arquitetura

O sistema é composto pelos seguintes componentes:

- **Server**: Serviço central que coordena a comunicação entre clientes e pontos de recarga
- **Charger**: Representa os pontos de recarga físicos
- **Client**: Simula os veículos elétricos

### Servidor (Server)

O servidor atua como intermediário entre os clientes (veículos) e os pontos de recarga. No arquivo `server.go`:

- Gerencia uma lista de pontos de recarga disponíveis no sistema
- Processa requisições de clientes para listar pontos próximos baseado em coordenadas geográficas
- Coordena o processo de reserva de pontos de recarga
- Controla o início e fim das sessões de carregamento
- Calcula o valor a ser pago com base no tempo de carregamento
- Gerencia os pagamentos das sessões
- Implementa lógica para calcular distâncias entre veículos e pontos de recarga

O servidor utiliza mutex para garantir operações thread-safe em dados compartilhados como o mapa de carros em carregamento.

### Ponto de Recarga (Charger)

Os pontos de recarga gerenciam o acesso físico às estações de carregamento. No arquivo `charger.go`:

- Mantém sua posição geográfica (latitude e longitude)
- Gerencia uma fila de espera de veículos
- Responde a solicitações do servidor para informações sobre o ponto
- Adiciona veículos à fila quando solicitado
- Verifica a prioridade dos veículos (se está primeiro na fila)
- Finaliza sessões de recarga
- Utiliza mutex para proteger o acesso concorrente à fila de espera

Cada ponto de recarga possui um ID único e opera em uma porta TCP específica, permitindo comunicação direta com o servidor.

### Fluxo do Sistema

1. Veículos monitoram seu nível de bateria
2. Quando a bateria está baixa, o veículo solicita uma lista de pontos de recarga próximos
3. O veículo seleciona um ponto e faz uma reserva
4. Após a autorização, o veículo inicia o carregamento
5. Ao finalizar, o sistema calcula o valor a ser pago
6. O veículo pode pagar a pendência posteriormente

## Estrutura do Projeto

```
.
├── Makefile                  # Comandos para build e execução
├── docker-compose.yml        # Configuração dos contêineres
├── charger/                  # Implementação dos pontos de recarga
│   ├── Dockerfile
│   ├── charger.go
│   └── go.mod
├── client/                   # Implementação dos clientes (veículos)
│   ├── Dockerfile
│   ├── client.go
│   └── go.mod
└── server/                   # Implementação do servidor central
    ├── Dockerfile
    ├── server.go
    └── go.mod
```

## Funcionalidades

- **Monitoramento Automático**: Clientes simulam o consumo de bateria e solicitam recarga quando necessário
- **Localização Geográfica**: Pontos de recarga possuem coordenadas geográficas e o sistema calcula distâncias
- **Gerenciamento de Filas**: Cada ponto de recarga mantém uma fila de espera para veículos
- **Processamento de Pagamentos**: Cálculo automático do valor baseado no tempo de carregamento
- **Histórico de Sessões**: Clientes mantêm um histórico de sessões de carregamento

## Protocolo de Comunicação

A comunicação entre os componentes é feita através de mensagens JSON com a seguinte estrutura:

```json
{
  "action": "NOME_DA_AÇÃO",
  "content": {
    "chave1": "valor1",
    "chave2": "valor2"
  }
}
```

### Ações Principais

- `LISTAR_PONTOS`: Solicita lista de pontos de recarga disponíveis
- `RESERVAR_PONTO`: Solicita reserva em um ponto específico
- `INICIO_CARREGAMENTO`: Inicia o processo de carregamento
- `FIM_CARREGAMENTO`: Finaliza o processo de carregamento
- `PAGAR_PENDENCIA`: Realiza pagamento de uma sessão

## Requisitos

- Docker e Docker Compose
- Go 1.20 ou superior (para desenvolvimento)

## Executando o Sistema

Para executar o sistema completo:

1. Construa todas as imagens:
   ```
   make build
   ```

2. Execute o servidor:
   ```
   make up-server
   ```

3. Em outro terminal, execute um ponto de recarga:
   ```
   make up-charger
   ```

4. Em outro terminal, execute outro ponto de recarga:
   ```
   make up-charger2
   ```

5. Execute um cliente (veículo):
   ```
   make run-client
   ```

6. Execute outro cliente (veículo):
   ```
   make run-client1
   ```

Para encerrar todos os contêineres:
```
make down
```

Para limpar imagens e volumes:
```
make clean
```

## Interface do Cliente (Veículo)

O cliente possui uma interface interativa com as seguintes opções:

- `B` - Simula bateria crítica e solicita lista de pontos de recarga
- `R` - Reserva um ponto de recarga
- `I` - Inicia carregamento
- `F` - Finaliza carregamento
- `P` - Paga última pendência
