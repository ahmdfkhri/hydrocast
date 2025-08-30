server:
	cd backend &&\
	go build -o bin/server cmd/server/main.go &&\
	./bin/server &&\
	cd ..

seed:
	cd backend &&\
	go build -o bin/seed cmd/seed/main.go &&\
	./bin/seed &&\
	cd ..

tidy:
	cd backend && go mod tidy && cd ..

gen: gen-go

GRPC_GO_OUT_DIR=backend/pkg
gen-go:
	protoc \
	--go_out=$(GRPC_GO_OUT_DIR) \
	--go-grpc_out=$(GRPC_GO_OUT_DIR) \
	protos/*.proto

TLS_CERT_DIR=backend/config/x509

dev-cert:
	openssl req -newkey rsa:2048 -nodes -keyout $(TLS_CERT_DIR)/server.key -out $(TLS_CERT_DIR)/server.csr && \
	openssl x509 -signkey $(TLS_CERT_DIR)/server.key -in $(TLS_CERT_DIR)/server.csr -req -days 365 -out $(TLS_CERT_DIR)/server.crt