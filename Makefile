

pb:
	@protoc -I=./ --go_out=./ --go_opt=paths=source_relative ./net/packet/packet.proto

test:
	go test ./...
