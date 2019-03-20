TEMPLATE_FILE=template.yml
FUNCTIONS=build/main

clean:
	rm -rf build sam.yml

build/helper: helper/*.go
	go build -o build/helper ./helper

build/main: *.go
	env GOARCH=amd64 GOOS=linux go build -o build/main

sam.yml: $(FUNCTIONS) template.yml build/helper
	aws cloudformation package \
		--template-file $(TEMPLATE_FILE) \
		--s3-bucket $(shell ./build/helper get CodeS3Bucket) \
		--s3-prefix $(shell ./build/helper get CodeS3Prefix) \
		--output-template-file sam.yml

deploy: sam.yml build/helper
	aws cloudformation deploy \
		--template-file sam.yml \
		--stack-name $(shell ./build/helper get StackName) \
		--capabilities CAPABILITY_IAM $(shell ./build/helper mkparam)
