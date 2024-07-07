test-it:
	go test ./...

# just reusing stage environment variables for convience
# include .env
# Manually created the cloudflare dns record local.strengthgadget.com
b:
	mkdir -p /tmp/strengthgadget
	GOARCH=arm64 GOOS=linux go build -o /tmp/strengthgadget/app .
	ssh "piegarden@173.197.226.162" "mkdir -p local.strengthgadget.com/session_data && cp staging.strengthgadget.com/.env local.strengthgadget.com/ && sudo docker compose -p local stop app"
	scp docker-compose-local.yaml keydb.conf /tmp/strengthgadget/app "piegarden@173.197.226.162:local.strengthgadget.com"
	ssh "piegarden@173.197.226.162" "cd local.strengthgadget.com && sudo docker compose -f docker-compose-local.yaml -p local up -d && sudo docker compose -p local start app"

logs:
	ssh "piegarden@173.197.226.162" "sudo docker compose -p local logs app"

redis-exec:
	sudo docker compose -p local exec keydb keydb-cli
