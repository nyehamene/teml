@run-src TARGET:
	go run ./cmd/eml/main.go generate -f ./codegen/go/generator/test-component/{{TARGET}}/source.teml
	go test ./codegen/go/generator/test-component/{{TARGET}}

@test:
	just run-src text-element
	just run-src text-group-element
