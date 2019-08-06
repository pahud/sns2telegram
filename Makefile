HANDLER ?= main
PACKAGE ?= $(HANDLER)
GOPATH  ?= $(HOME)/go
GOOS    ?= linux
GOOSDEV	?= $(shell uname -s)
GOARCH  ?= amd64
S3BUCKET	?= pahud-tmp-ap-northeast-1
STACKNAME	?= sns2telegram
LAMBDA_REGION ?= ap-northeast-1
LAMBDA_FUNC_NAME ?= sns2telegram

WORKDIR = $(CURDIR:$(GOPATH)%=/go%)
ifeq ($(WORKDIR),$(CURDIR))
	WORKDIR = /tmp
endif

all: dep build pack package

dep:
	@echo "Checking dependencies..."
	@dep ensure

build:
ifeq ($(GOOS),darwin)
	@docker run -ti --rm -v $(shell pwd):/go/src/myapp.github.com -w /go/src/myapp.github.com  golang:1.10 /bin/sh -c "make build-darwin"
else
	@docker run -ti --rm -v $(shell pwd):/go/src/myapp.github.com -w /go/src/myapp.github.com  golang:1.10 /bin/sh -c "make build-linux"
endif

run:
	@docker run -ti --rm -v $(shell pwd):/go/src/myapp.github.com -w /go/src/myapp.github.com  golang:1.10 /bin/sh -c "go run *.go"


build-linux:
	@GOOS=linux GOARCH=amd64 go build -o main
	# @go get -u github.com/golang/dep/cmd/dep
	# @[ ! -f ./Gopkg.toml ] && dep init || true
	# @dep ensure
	# @GOOS=linux GOARCH=amd64 go build -o main

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o main
	# @go get -u github.com/golang/dep/cmd/dep
	# @[ ! -f ./Gopkg.toml ] && dep init || true
	# @dep ensure
	# @GOOS=darwin GOARCH=amd64 go build -o main


# build:
# 	@echo "Building..."
# 	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags='-w -s' -o $(HANDLER)

# devbuild:
# 	@echo "Building..."
# 	@GOOS=$(GOOSDEV) GOARCH=$(GOARCH) go build -ldflags='-w -s' -o $(HANDLER)

# pack:
# 	@echo "Packing binary..."
# 	@zip $(PACKAGE).zip $(HANDLER)

# clean:
# 	@echo "Cleaning up..."
# 	@rm -rf $(HANDLER) $(PACKAGE).zip

package:
	@echo "sam packaging..."
	@aws cloudformation package --template-file sam.yaml --s3-bucket $(S3BUCKET) --output-template-file sam-packaged.yaml

deploy:
	@echo "sam deploying..."
	@aws cloudformation deploy --template-file sam-packaged.yaml --stack-name $(STACKNAME) --capabilities CAPABILITY_IAM

	
.PHONY: func-prep	
func-prep:
	@rm -f main.zip; zip -r main.zip main
	
.PHONY: sam-package
sam-package:
	@docker run -ti \
	-v $(PWD):/home/samcli/workdir \
	-v $(HOME)/.aws:/home/samcli/.aws \
	-w /home/samcli/workdir \
	-e AWS_DEFAULT_REGION=$(LAMBDA_REGION) \
	pahud/aws-sam-cli:latest sam package --template-file sam.yaml --s3-bucket $(S3BUCKET) --output-template-file packaged.yaml


.PHONY: sam-package-from-sar
sam-package-from-sar:
	@docker run -ti \
	-v $(PWD):/home/samcli/workdir \
	-v $(HOME)/.aws:/home/samcli/.aws \
	-w /home/samcli/workdir \
	-e AWS_DEFAULT_REGION=$(LAMBDA_REGION) \
	pahud/aws-sam-cli:latest sam package --template-file sam-sar.yaml --s3-bucket $(S3BUCKET) --output-template-file packaged.yaml

.PHONY: sam-publish
sam-publish:
	@docker run -ti \
	-v $(PWD):/home/samcli/workdir \
	-v $(HOME)/.aws:/home/samcli/.aws \
	-w /home/samcli/workdir \
	-e AWS_DEFAULT_REGION=$(LAMBDA_REGION) \
	pahud/aws-sam-cli:latest sam publish --region $(LAMBDA_REGION) --template packaged.yaml
	
	
.PHONY: sam-publish-global
sam-publish-global:
	$(foreach LAMBDA_REGION,$(GLOBAL_REGIONS), LAMBDA_REGION=$(LAMBDA_REGION) make sam-publish;)


.PHONY: sam-deploy
sam-deploy:
	@docker run -ti \
	-v $(PWD):/home/samcli/workdir \
	-v $(HOME)/.aws:/home/samcli/.aws \
	-w /home/samcli/workdir \
	-e AWS_DEFAULT_REGION=$(LAMBDA_REGION) \
	pahud/aws-sam-cli:latest sam deploy \
	--parameter-overrides FunctionName=$(LAMBDA_FUNC_NAME)  \
	--template-file ./packaged.yaml --stack-name "$(LAMBDA_FUNC_NAME)" --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND
	# print the cloudformation stack outputs
	aws --region $(LAMBDA_REGION) cloudformation describe-stacks --stack-name "$(LAMBDA_FUNC_NAME)" --query 'Stacks[0].Outputs'


.PHONY: sam-logs-tail
sam-logs-tail:
	@docker run -ti \
	-v $(PWD):/home/samcli/workdir \
	-v $(HOME)/.aws:/home/samcli/.aws \
	-w /home/samcli/workdir \
	-e AWS_DEFAULT_REGION=$(LAMBDA_REGION) \
	pahud/aws-sam-cli:latest sam logs --name $(LAMBDA_FUNC_NAME) --tail

.PHONY: sam-destroy
sam-destroy:
	# destroy the stack	
	aws --region $(LAMBDA_REGION) cloudformation delete-stack --stack-name "$(LAMBDA_FUNC_NAME)"

