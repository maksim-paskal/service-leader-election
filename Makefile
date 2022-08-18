KUBECONFIG=$(HOME)/.kube/dev
image=paskalmaksim/service-leader-election
tag=dev

test:
	./scripts/validate-license.sh
	go fmt ./cmd/... ./pkg/...
	go vet ./cmd/... ./pkg/...
	go mod tidy
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -v

lint:
	ct lint --charts ./chart

build:
	go run github.com/goreleaser/goreleaser@latest build --rm-dist --snapshot --skip-validate
	mv dist/service-leader-election_linux_amd64_v1/service-leader-election ./service-leader-election
	docker buildx build --pull --push . -t $(image):$(tag)

run:
	go run -race ./cmd/main.go -kubeconfig=$(KUBECONFIG)

deploy:
	kubectl -n service-leader-election scale deploy service-leader-election --replicas=0 || true
	helm upgrade service-leader-election ./chart \
	--install \
	--create-namespace \
	--namespace service-leader-election

restart:
	kubectl -n service-leader-election rollout restart deployment/service-leader-election

clean:
	kubectl delete ns service-leader-election