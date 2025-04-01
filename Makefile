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

down-client:
	docker-compose stop client
	docker-compose rm -f client

# Remover imagens e volumes antigos
clean:
	docker system prune -a -f

# Roda o client
build-client:
	docker-compose build client

run-client:
	docker-compose run client


