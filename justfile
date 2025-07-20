@run-src TARGET:
	go run ./cmd/eml/main.go generate -f ./codegen/go/generator/test-component/{{TARGET}}/source.teml
	go test -cover -timeout 3s ./codegen/go/generator/test-component/{{TARGET}}

@test-all:
	just run-src text-element
	just run-src text-group-element
	just run-src string-element
	just run-src string-element-attr
	just run-src number-element
	just run-src number-element-attr
	just run-src component-element
	just run-src component-element-attr

@test TARGET:
	just run-src {{TARGET}}
