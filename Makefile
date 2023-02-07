build:
	swag init -g cmd/app/main.go
	go build -o apiproxychecker cmd/app/main.go

run:
	swag init -g cmd/app/main.go
	go run cmd/app/main.go
