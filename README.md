# Devroute contract reference implementation

This repo contains a reference implementation of Devroute, a contract that we implemented at Omio
with the purpose of making easier to run end-to-end tests. For more context, please read our blog post. 
You can use this repo to learn more about how the contract works, and in general, as an example of how an [Istio's EnvoyFilter](https://istio.io/v1.5/docs/reference/config/networking/envoy-filter/) can be used for a real-life use case. 

This repo contains 3 folders:

- **demo**: contains everything necessary to run a demo of this contract. 
- **devrouter**: reference implementation of a simple proxy that reads the devroute contract and proxies 
requests according to the address specified in the contract.
- **envoy-filter**: contains an implementation of `envoy_on_request` function that fulfills the devroute 
contract. 
