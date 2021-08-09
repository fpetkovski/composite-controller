.PHONY: generate
generate:
	controller-gen crd paths=./pkg/apis/... output:stdout > deploy/crds.yaml
	controller-gen object paths=./pkg/apis/...

.PHONY: kind-cluster
kind-cluster:
	kind create cluster --config hack/kind.yaml || true
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

.PHONY: format
format:
	goimports -l -w .

.PHONY: test-e2e
test-e2e:
	go test ./test/e2e/... -v