all: container

TAG = 0.0
PREFIX = registry.lpc-win32.com/common/hpa-custom-controller

server:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w'
	go bulid -o hpactl/hpactl hpactl/hpactl.go

container: server
	docker build -t $(PREFIX):$(TAG) . --no-cache
	docker push $(PREFIX):$(TAG)

clean:
	rm -f hpa-custom-controller
