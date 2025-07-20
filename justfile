@test TARGET:
	go run ./cmd/eml/main.go generate -f ./codegen/go/generator/test-component/{{TARGET}}/source.teml
	go test -cover -timeout 3s ./codegen/go/generator/test-component/{{TARGET}}

@test-all:
	just test text-element
	just test text-group-element
	just test string-element
	just test string-element-attr
	just test number-element
	just test number-element-attr
	just test component-element
	just test component-element-attr
	just test native-element
	just test native-element-attr

