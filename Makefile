.ONESHELL:

PACKAGE_DIRS=$(shell go list ./... | grep -v /vendor/)
ECR_URI="620055013658.dkr.ecr.us-west-2.amazonaws.com"

tidy:
	@go mod tidy

test:
	@echo "crossed fingers emoji running tests"
	@go test -v $(PACKAGE_DIRS)

# XXX: migrate to kubernetes secrets

secrets:
	yq . ${HOME}/.bs/secrets.yml > etc/bs/secrets.yml
	yq . ${HOME}/.sux/secrets.yml > etc/sux/secrets.yml
	@echo "local secrets file is now tainted, use \"make rmsecrets\" to remove before committing"

rmsecrets:
	@echo "removing local secrets"
	echo > etc/bs/secrets.yml
	echo > etc/sux/secrets.yml

dockerauth:
	aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin $(ECR_URI)

docker:
	docker buildx build --no-cache --tag $(ECR_URI):latest --platform linux/amd64 --push .

kubelogs:
	kubectl logs `kubectl get pods | grep Running | grep -i becky | cut -d ' ' -f 1`

bounce:
	kubectl rollout restart deployment archeavy-becky

version:
	@echo "Updating version data"
	@echo "version:" > etc/bs/version.yml
	@echo "  build_date: \"`date`\"" >> etc/bs/version.yml
	@echo "  build: \"`git describe --tags --always`\"" >> etc/bs/version.yml
	@echo "  branch: \"`git branch | grep '^*' | cut -d' ' -f 2`\"" >> etc/bs/version.yml

# TODO: make local -- run locally
local:

# XXX: untested
fwd:
	kubectl port-forward pod/$(kubectl get pods | grep Running | grep -i becky | cut -d ' ' -f 1) 8080:8080
