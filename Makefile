

pb:
	@protoc -I=./ --go_out=./ --go_opt=paths=source_relative ./net/packet/packet.proto

gopb:
	@protoc -I=./ --gogofaster_out=. --gogofaster_opt=paths=source_relative ./network/packet/packet.proto


test:
	go test ./...


mock:
	mockgen ./net/process Context,PacketEncoder,PacketDecoder

gen: mock
	go generate ./...
