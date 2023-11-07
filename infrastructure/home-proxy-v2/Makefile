# First execute with sudo:  sudo mkdir -p /var/shared && sudo chmod 777 /var/shared
deploy:
	cp ./certbot/conf/cloudflare.ini /var/shared/cloudflare.ini # This is so we don't have to deal with relative paths
	docker compose build
	docker compose up -d haproxy watcher

update-config:
	mkdir -p /usr/local/etc/haproxy
	cp ./haproxy.cfg /usr/local/etc/haproxy/
