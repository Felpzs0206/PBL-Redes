.PHONY: build up down clean client

# Subir o servidor e o charger com build

build:
	docker-compose build server charger charger2 client client1
# Subir os serviços sem forçar rebuild
build-server:
	docker-compose build server
build-charger:
	docker-compose build charger charger2
build-client:
	docker-compose build client client1
up-server:
	docker-compose up server

up-charger:
	docker-compose up charger

up-charger2:
	docker-compose up charger2

run-client:
	docker-compose run client 
run-client1:
	docker-compose run client1

# Derrubar os contêineres
down:
	docker-compose down --volumes --remove-orphans

down-client:
	docker-compose stop client
	docker-compose rm -f client


# Remover imagens e volumes antigos
clean:
	docker system prune -a -f


