docker:
	docker build -t ncserver /home/t/pg/shellbin/nc-server
	docker image tag ncserver  docker-registry:5000/ncserver
	docker push docker-registry:5000/ncserver