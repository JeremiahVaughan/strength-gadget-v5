# RUN MANUALLY TO START CLUSTER
# kind delete cluster || true && kind create cluster --config kind/kind.yaml
deploy-local: init-database-and-user parallel final
deploy-local-debug: init-database-and-user linear final
init: context-kind init-database-cluster init-db-operator-client deploy-local

PARALLEL_STEPS := step1 step2 step3 step4 step5

parallel: context-kind
	@echo "Running parallel steps..."
	@$(MAKE) $(PARALLEL_STEPS:%=%-parallel) --jobs=5

linear: context-kind
	@echo "Running linear steps..."
	@$(MAKE) $(PARALLEL_STEPS:%=%-parallel) --jobs=1

step1-parallel: just-strengthgadget
step2-parallel: db-proxy
step3-parallel: test-tornado
step4-parallel: unit-tests
step5-parallel: load-balancer

final:
	echo final-step

# Reference: https://kind.sigs.k8s.io/docs/user/loadbalancer/
load-balancer:
	kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.7/config/manifests/metallb-native.yaml
	kubectl wait --namespace metallb-system \
                    --for=condition=ready pod \
                    --selector=app=metallb \
                    --timeout=90s
	kubectl apply -f https://kind.sigs.k8s.io/examples/loadbalancer/metallb-config.yaml

init-database-cluster: context-kind
	helm repo add haproxy-ingress https://haproxy-ingress.github.io/charts
	helm install haproxy-ingress haproxy-ingress/haproxy-ingress --create-namespace --namespace ingress-controller --version 0.14.2 -f kind/haproxy-ingress-values.yaml
	kubectl apply -f kind/haproxy-config.yaml
	kubectl apply -f https://raw.githubusercontent.com/cockroachdb/cockroach-operator/v2.10.0/install/crds.yaml
	kubectl apply -f https://raw.githubusercontent.com/cockroachdb/cockroach-operator/v2.10.0/install/operator.yaml
	kubectl apply -f kind/database-service.yaml
	$(call retry,kubectl apply -f kind/database.yaml)
init-db-operator-client: context-kind
	$(call retry,kubectl apply -f kind/database-operator.yaml)
init-database-and-user: context-kind
	$(call retry,kubectl exec -it cockroachdb-client-secure -- ./cockroach sql --certs-dir=/cockroach/cockroach-certs --host=cockroachdb-public --execute="DROP DATABASE IF EXISTS auth CASCADE;CREATE DATABASE auth;DROP USER IF EXISTS strengthgadget;CREATE USER strengthgadget WITH PASSWORD 'not_a_real_password_do_not_use_in_production';")

just-strengthgadget:
	kubectl delete -f kind/redis.yaml || true
	kubectl apply -f kind/redis.yaml
	kubectl delete -f kind/strengthgadget.yaml > /dev/null || true
	docker build -t localhost:5000/strengthgadget:0.0.1 . && kind load docker-image localhost:5000/strengthgadget:0.0.1 --name kind
	kubectl apply -f kind/strengthgadget.yaml

test-tornado:
	kubectl delete -f kind/local_secret.yaml > /dev/null || true
	kubectl delete -f kind/test_tornado.yaml > /dev/null || true
	docker build -t localhost:5000/test-tornado:0.0.1 -f test_tornado/Dockerfile . && kind load docker-image localhost:5000/test-tornado:0.0.1 --name kind
	kubectl apply -f kind/local_secret.yaml
	kubectl apply -f kind/test_tornado.yaml

get-ingress-ip: context-kind
	kubectl get service haproxy-ingress -o jsonpath='{.status.loadBalancer.ingress[0].ip}' -n ingress-controller

db-proxy: context-kind
	kubectl delete -f kind/haproxy-deployment.yaml > /dev/null || true
	docker build -f kind/Dockerfile-haproxy -t localhost:5000/haproxy:0.0.1 kind/. && kind load docker-image localhost:5000/haproxy:0.0.1 --name kind
	kubectl apply -f kind/haproxy-deployment.yaml

context-kind:
	#kubectl config use-context kind
	kubectl config use-context kind-kind
unit-tests:
	go test ./...

define retry
    @n=0; \
    max=120; \
    while [ $$n -lt $$max ]; do \
        $(1) && break || { \
            echo "Command failed"; \
            if [ $$n -lt $$max ]; then \
                n=$$((n+1)); \
                echo "Retrying ($$n/$$max)"; \
                sleep 3; \
            else \
                echo "Failed after $$n attempts"; \
                exit 1; \
            fi \
        }; \
    done
endef

