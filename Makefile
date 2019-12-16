TAG=vlkoti/mible

build:
	docker build -t $(TAG) .

push:
	docker push -t $(TAG)
