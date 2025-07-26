@test TARGET:
	go run ./cmd/eml/main.go generate -f ./codegen/go/generator/test-template/{{TARGET}}/source.teml
	go test -cover -timeout 3s ./codegen/go/generator/test-template/{{TARGET}}

@test-all:
	just test text-element
	just test text-group-element
	just test string-element
	just test string-element-attr
	just test number-element
	just test number-element-attr
	just test component-element
	just test component-element-attr
	just test component-element-body
	just test native-element
	just test native-element-attr
	just test if-element
	just test cond-element
	just test instance-element
	just test document-element
	just test document-tagged-attr

