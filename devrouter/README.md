# Devrouter

A simple proxy that parses the contents of `devroute` http header and uses that information 
to proxy requests. 

Devroute exists because [Envoy's Lua Filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter) makes http calls only to members of the mesh. 
