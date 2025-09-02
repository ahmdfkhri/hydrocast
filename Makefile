include .env

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

admin:
	cd backend &&\
	go build -o bin/admin cmd/admin/main.go &&\
	./bin/admin &&\
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

CA_KEY=$(TLS_CERT_DIR)/ca.key
CA_CRT=$(TLS_CERT_DIR)/ca.crt

SRV_KEY=$(TLS_CERT_DIR)/server.key
SRV_CSR=$(TLS_CERT_DIR)/server.csr
SRV_CRT=$(TLS_CERT_DIR)/server.crt
SRV_EXT=$(TLS_CERT_DIR)/server.v3.ext

ca:
	mkdir -p $(TLS_CERT_DIR)
	openssl genrsa -out $(CA_KEY) 4096
	openssl req -x509 -new -key $(CA_KEY) -sha256 -days 1826 -out $(CA_CRT) \
		-subj '/C=$(TLS_C)/ST=$(TLS_ST)/L=$(TLS_L)/O=$(TLS_O)/OU=$(TLS_OU)/CN=$(TLS_CN)'

$(SRV_EXT):
	@echo "authorityKeyIdentifier=keyid,issuer" > $(SRV_EXT)
	@echo "basicConstraints=CA:FALSE" >> $(SRV_EXT)
	@echo "keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment" >> $(SRV_EXT)
	@echo "subjectAltName = @alt_names" >> $(SRV_EXT)
	@echo "" >> $(SRV_EXT)
	@echo "[alt_names]" >> $(SRV_EXT)
	@echo "DNS.1 = localhost" >> $(SRV_EXT)
	@echo "IP.1 = 127.0.0.1" >> $(SRV_EXT)

dev-server-cert: $(SRV_EXT)
	openssl req -new -nodes -out $(SRV_CSR) -newkey rsa:4096 -keyout $(SRV_KEY) \
		-subj '/C=$(TLS_C)/ST=$(TLS_ST)/L=$(TLS_L)/O=$(TLS_O)/OU=$(TLS_OU)/CN=$(TLS_CN)'

	openssl x509 -req -in $(SRV_CSR) -CA $(CA_CRT) -CAkey $(CA_KEY) -CAcreateserial \
		-out $(SRV_CRT) -days 730 -sha256 -extfile $(SRV_EXT)