test-it:
	go test ./...

# just reusing stage environment variables for convience
# include .env
# Manually created the cloudflare dns record local.strengthgadget.com
b:
	go run ./...

logs:
	ssh "piegarden@173.197.226.162" "sudo docker compose -p local logs app"

logs-staging:
	ssh "piegarden@173.197.226.162" "sudo docker compose -p staging logs app"

logs-production:
	ssh "piegarden@173.197.226.162" "sudo docker compose -p production logs app"
