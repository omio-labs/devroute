# FAQ

Lua Envoy Filter contract requires users to define custom routing logic via two functions:
 - `envoy_on_request(request_handle)`
 - `envoy_on_respond(request_handle)`

each function receives a `request_handle` which is created by Envoy internally. We 
can only call functions on that object, we can't create instances of it. 

Given the above conditions, to be able to test our implementation we need to mock the
`request_handle` object.

**Can't we use Lua modules instead of hacking with `loadfile` function ?**

If we were to own Istio configuration, we could mount a volume on every Envoy container 
and drop a Lua module that we can load in the EnvoyFilter. This would avoid having to do 
the workaround of using `loadfile` function to allow `filter_test` to test the functions defined in 
`filter.lua`. With a module, we would just import the module from the test and not depend on this hack. 
Since GKE owns Istio configuration, we can't change Envoy container PodSpec.




