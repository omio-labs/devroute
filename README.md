# Devroute reference implementation

This repo contains a reference implementation of Devroute, a contract that we implemented at Omio
with the purpose of making easier to run end-to-end tests. For more context, please read our blog post. 
You can use this repo to learn more about how the contract works, and in general, as an example of how an [Istio's EnvoyFilter](https://istio.io/v1.5/docs/reference/config/networking/envoy-filter/) can be used for a real-life use case. 

# How to run

## Prerequisites

- A Kubernetes cluster running with Istio. This implementation has been tested with: 
    - k8s 1.16
    - istio 1.4
- Docker

## Deploy resources in kubernetes

```
kubectl apply -f foo.yaml
kubectl apply -f envoyfilter.yaml
kubectl apply -f devrouter.yaml
```

we will be now making a call to the `foo` pod running in kubernetes. We can either use the 
nodeport or ingress. In this example, we will use the nodeport

```
curl http://<a-k8s-node-ip>:30111/get
```

now run another replica of `foo` on your laptop. We can simply run this on docker:

```
docker run -it -p 8001:80 kennethreitz/httpbin:latest gunicorn --access-logfile - -b 0.0.0.0:80 httpbin:app
```

Let's detour traffic to our docker version of `foo` by adding the contract header. 
You would need to know your laptop's IP. 

Let's make the request:

```
curl -H 'x-devroute: {"foo":"<laptops-ip>:8001"}' http://<a-k8s-node-ip>:30111/get
```

This request should be now hitting the `foo` running in docker instead of the one running 
in our kubernetes cluster. To convince yourself, you can check the logs both on the docker 
instance of `foo` and on the pod in kubernetes. You will see that the last called was rerouted. 
