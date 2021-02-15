-- The service name in which this filter is running.
-- Not to be confused with k8s service, it's rather the common 
-- name that developers use to name whatever this pod is running.
-- This global variable has to be replaced when applying the 
-- envoy filter on a given namespace. 
-- At Omio, we use a mutating webhook to do this.
-- If you have full control of Istio installation, you can simply 
-- modify Istio's side-car PodSpec to contain an environment variable 
-- that can be then read from this code. 
SERVICE_NAME = "devroute_placeholder_service_name"

-- get_req_body extracts the body from request_handle 
function get_req_body(request_handle)
    local req_body_buf = request_handle:body() 
    if req_body_buf ~= nil then
        return req_body_buf:getBytes(0, req_body_buf:length())
    end
    return ""
end

-- escape_lua_magic escapes str in case it has 
-- lua magic characters. 
function escape_lua_magic(str)
    -- lua magic characters:
    -- (   )   .   %   +   –   *   ?   [   ^   $
    local pattern = "[%(%)%.%%%+%-%*%?%[%^%$]"
    local replacements = {
        ['('] = "%(",
        [')'] = "%)",
        ['.'] = "%.",
        ['%'] = "%%",
        ['+'] = "%+",
        ['-'] = "%-",
        ['*'] = "%*",
        ['?'] = "%?",
        ['['] = "%[",
        ['^'] = "%^",
        ['$'] = "%$"
    }
    return string.gsub(str, pattern, replacements)
end

-- envoy_on_request checks for the presence of `x-devroute` contract header 
-- and detours this request to devrouter if there is a service match. A service
-- specified in the contract header matches this filter if it's equal to SERVICE_NAME.
function envoy_on_request(request_handle)
    local contract = request_handle:headers():get("x-devroute")
    if contract == nil then
        return
    end

    -- check if service name as specified in contract header matches the service name this 
    -- EnvoyFilter is running in.
    if string.match(contract, '.*["\']' .. escape_lua_magic(SERVICE_NAME) .. '["\'].*') == nil then
        return 
    end

    -- add matched header so devrouter can pick the right ip:host among 
    -- multiple services specified in contract.
    request_handle:headers():add("x-devroute-matched", SERVICE_NAME)
    -- preserve request headers
    local in_headers = {}
    for key, value in pairs(request_handle:headers()) do
        in_headers[key] = value
    end
    
    -- call devrouter
    in_headers["authority"] = "devrouter.devrouter.svc.cluster.local"
    local resp_headers, resp_body = request_handle:httpCall(
        "outbound|80||devrouter.devrouter.svc.cluster.local", 
        in_headers,
        get_req_body(request_handle),
        5000)
    request_handle:respond(resp_headers, resp_body)
end
