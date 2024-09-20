docker:
	docker build -t rustwebserver /home/t/pg/shellbin/db-service
	docker image tag db-service docker-registry:5000/db-service
	docker push docker-registry:5000/db-service
