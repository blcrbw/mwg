
gall:
	kubectl get all

up:
	helm repo add hashicorp https://helm.releases.hashicorp.com
	helm install consul hashicorp/consul --set global.name=consul --create-namespace --namespace consul
	#kubectl port-forward service/consul-server 8500:8500 -n consul

down:
	helm uninstall consul
	sleep 5
	kubectl get all -o wide

c-up:
	k3d cluster create mwg --servers 3 --agents 0 --registry-create mwg-registry:5001

c-start:
	k3d cluster start mwg

c-stop:
	k3d cluster stop mwg

c-down:
	k3d cluster delete mwg

build-rating:
	GOOS=linux go build -o rating/cmd/main rating/cmd/main.go
	docker build -t rating:latest ./rating/
	docker tag rating:latest localhost:5001/rating:latest
	docker push localhost:5001/rating:latest
	#k3d image import rating:latest -c mwg

build-metadata:
	GOOS=linux go build -o metadata/cmd/main metadata/cmd/main.go
	docker build -t metadata:latest ./metadata/
	docker tag metadata:latest  localhost:5001/metadata:latest
	docker push localhost:5001/metadata:latest --disable-content-trust
	#k3d image import metadata:latest -c mwg

build-movie:
	GOOS=linux go build -o movie/cmd/main movie/cmd/main.go
	docker build -t movie:latest ./movie/
	docker tag movie:latest localhost:5001/movie:latest
	docker push localhost:5001/movie:latest
	#k3d image import movie:latest -c mwg

kube-deployment:
	kubectl apply -f rating/kubernetes-deployment.yaml
	kubectl apply -f movie/kubernetes-deployment.yaml
	kubectl apply -f metadata/kubernetes-deployment.yaml

tls-cert:
	openssl req -x509 -nodes -newkey rsa:4096 \
	-keyout cert.key -out cert.crt -days 365 -nodes \
	-subj "/C=US/ST=State/L=City/O=Organization/OU=Department/CN=localhost" \
	-addext "subjectAltName=DNS:localhost,DNS:example.com,DNS:movie,DNS:rating,DNS:metadata,IP:127.0.0.1,IP:192.168.1.1,IP:172.21.0.8,IP:172.21.0.6,IP:172.21.0.7" -config /dev/null
	cp cert.* metadata/configs/
	cp cert.* rating/configs/
	cp cert.* movie/configs/

make put-metadata:
	bash -c 'grpcurl -cacert <(cat cert.crt) -d '\''{"metadata": {"id":"the-movie", "title": "The Movie", "description": "", "director": "Mr. D"} }'\'' localhost:8081 MetadataService/PutMetadata'

make get-movie:
	bash -c 'grpcurl -cacert <(cat cert.crt) -d '\''{"movie_id":"the-movie"}'\'' localhost:8083 MovieService/GetMovieDetails'