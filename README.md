# Ingress SG Validator

## How to deploy

```
$ make docker-build
$ kind load docker-image controller:latest # TODO: move to Makefile
$ kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.5.3/cert-manager.yaml
$ make deploy
```

## How to run test

```
$ make test
```

## How to publish

```
$ IMG=naari3/ingress-sg-validator:latest make docker-build
$ IMG=naari3/ingress-sg-validator:latest make docker-push
```