@gen TARGET:
	go run ./cmd/tel/main.go generate -f ./cmd/internal/gentest/tests/{{TARGET}}/source.teml

@gen-all:
	just gen text-element
	just gen text-group-element
	just gen string-element
	just gen string-element-attr
	just gen number-element
	just gen number-element-attr
	just gen property-element
	just gen property-element-attr
	just gen property-element-body
	just gen native-element
	just gen native-element-attr
	just gen if-element
	just gen cond-element
	just gen component-element
	just gen document-element

@gen-test TARGET:
	go run ./cmd/internal/gentest/cmd.go ./cmd/internal/gentest/tests/{{TARGET}}/source.teml

@gen-test-all:
	just gen-test text-element
	just gen-test text-group-element
	just gen-test string-element
	just gen-test string-element-attr
	just gen-test number-element
	just gen-test number-element-attr
	just gen-test property-element
	just gen-test property-element-attr
	just gen-test property-element-body
	just gen-test native-element
	just gen-test native-element-attr
	just gen-test if-element
	just gen-test cond-element
	just gen-test component-element
	just gen-test document-element

@test-generated:
	go test -cover -timeout 3s ./cmd/internal/gentest/generated-tests/...

@test TARGET:
	go test -cover -timeout 3s {{TARGET}}

@test-all:
	just test ./token
	just test ./ast
	just test ./transpiler
	just test ./codegen/go/transpiler
	just test ./codegen/go/generator
	just test ./
