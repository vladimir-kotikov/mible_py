TAG=vlkoti/mible:go

image:
	docker build -t $(TAG) .

push: image
	docker push $(TAG)
