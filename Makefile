build:
	GOOS=linux GOARCH=amd64 go build -o api/rest/bootstrap api/rest/main.go
	GOOS=linux GOARCH=amd64 go build -o api/graphql/bootstrap api/graphql/main.go

zip:
	zip -j api/rest/main.zip api/rest/bootstrap
	zip -j api/graphql/main.zip api/graphql/bootstrap

deploy: build zip
	serverless deploy --config serverless.rest.yml	
	serverless deploy --config serverless.graphql.yml
	
remove:
	serverless remove --config serverless.rest.yml	
	serverless remove --config serverless.graphql.yml