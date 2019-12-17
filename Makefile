TAG=vlkoti/mible

image:
	docker build -t $(TAG) .

push: image
	docker push $(TAG)
