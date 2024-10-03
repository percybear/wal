# docker hub commands

1. login to docker hub locally
2. build the docker image
3. tag the docker image
4. push the docker image to docker hub
5. pull the docker image from docker hub

## step 1: Login to Docker Hub
```console
git docker login
```

## step 2 & 3: build and tag the docker image
```console
docker build -t github.com/pmoth/wal:0.0.2 .
```

## step 4: push the docker image to docker hub
```console
docker tag github.com/pmoth/wal:$(TAG) docker.io/pmoth/wal:$(TAG)
docker push docker.io/pmoth/wal:$(TAG)
```

## step 5: pull the docker image from docker hub
```console
docker pull docker.io/pmoth/wal:0.0.2
```

## step 6: run the docker image
```console
docker run -it docker.io/pmoth/wal:0.0.2
```

