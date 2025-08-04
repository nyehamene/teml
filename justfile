@gen-test TARGET:
	go run ./cmd/eml/main.go generate -f ./codegen/go/generator/test-template/{{TARGET}}/source.teml
	go test -cover -timeout 3s ./codegen/go/generator/test-template/{{TARGET}}

@gen-test-all:
	just gen-test text-element
	just gen-test text-group-element
	just gen-test string-element
	just gen-test string-element-attr
	just gen-test number-element
	just gen-test number-element-attr
	just gen-test component-element
	just gen-test component-element-attr
	just gen-test component-element-body
	just gen-test native-element
	just gen-test native-element-attr
	just gen-test if-element
	just gen-test cond-element
	just gen-test instance-element
	just gen-test document-element

@test TARGET:
	go test -cover -timeout 3s {{TARGET}}

@test-all:
	just test ./token
	just test ./ast
	just test ./transpiler
	just test ./codegen/go/transpiler
	just test ./codegen/go/generator
	just test ./

