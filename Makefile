.PHONY: build up down clean client

# Subir o servidor e o charger com build
build:
	docker-compose build server charger client
# Subir os serviços sem forçar rebuild
build-server:
	docker-compose build server
build-charger:
	docker-compose build charger
build-client:
	docker-compose build client
up-server:
	docker-compose up server

up-charger:
	docker-compose up charger

run-client:
	docker-compose run client

# Derrubar os contêineres
down:
	docker-compose down --volumes --remove-orphans

down-client:
	docker-compose stop client
	docker-compose rm -f client


# Remover imagens e volumes antigos
clean:
	docker system prune -a -f


