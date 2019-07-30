
SRC_DIR = result
DST_DIR = result
RESULT_PROTO = result.proto

SERVER_IP = 0.0.0.0
PORT = 5001


build:
	 protoc -I $(SRC_DIR)/ $(SRC_DIR)/${RESULT_PROTO} --go_out=plugins=grpc:result \
	&& go build

master: 
	./goloadtest MASTER $(SERVER_IP) $(PORT) masterclient 
