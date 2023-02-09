build:
	swag init -g cmd/app/main.go
	go build -o proxychecker cmd/app/main.go

run:
	swag init -g cmd/app/main.go
	go run cmd/app/main.go

d.build:
	swag init -g cmd/app/main.go
	docker build . -t proxychecker:v1

d.run:
	docker run --name proxychecker -p 30122:30122 proxychecker:v1

d.start:
	docker start proxychecker

d.c.build:
	docker-compose build

d.c.run:
	docker-compose up