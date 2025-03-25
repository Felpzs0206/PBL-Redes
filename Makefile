.PHONY: build up down clean client

# Subir o servidor e o charger com build
build:
	docker-compose up --build server charger

# Subir os serviços sem forçar rebuild
up:
	docker-compose up server charger

# Derrubar os contêineres
down:
	docker-compose down --volumes --remove-orphans

# Remover imagens e volumes antigos
clean:
	docker system prune -a -f

# Roda o client
client:
	docker-compose up --build client