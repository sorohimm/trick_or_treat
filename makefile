build:
	go	mod vendor && cd deployments && docker-compose -p balance_api build

run:
	cd deployments && docker-compose up

down:
	cd deployments && docker-compose down