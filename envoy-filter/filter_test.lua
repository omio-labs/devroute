require 'busted.runner'()

-- load filter.lua functions in this file.
-- workaround for module restriction (see README)
f = loadfile("/tests/filter.lua")
f()


-- NewRequestHandlerMock create mocks of request_handle. 
-- request_handle is the object that envoy instanciates internally 
-- and pass as parameter to envoy_on_request function.
-- A table in Lua can be considered an object, so here 
-- we add keys mapping to functions (object's method)
-- These functions have return values that can be customized 
-- for testing purposes via config param. 
function NewRequestHandlerMock(config)
    -- default empty config in case not specified
    if config == nil then 
        config = {}
    end
    
    -- headerCollection is a table that is used to store headers 
    -- when we call request_handle:headers():get('foo') or 
    -- request_handle:headers():add('foo', 'bar'). It's passed
    -- as a closure to headerMock.
    local headerCollection = config.request_headers or {}
    -- mock header object/functions
    local headerMock = {
        hvalues = headerCollection,
        get = function(self, k) return headerCollection[k] end,
        add = function(self, k, v)  headerCollection[k] = v end
    }
    -- mock body object/functions
    local bodyMock = {
        getBytes = function(self, index, length) return config.request_body or "" end,
        length = function(self) return 42 end   
    }
    -- request_handle mock
    return {
        headers = function(self) return headerMock end,
        body = function(self) return bodyMock end,
        httpCall = 
        function(self, cluster, headers, body, timeout) 
            return config.response_headers, config.response_body 
        end,
        respond = function(self, headers, body) end
    }
end

describe("envoy_on_request", function()
    setup(function()
        match = require("luassert.match")
        _ = match._ 
        -- backup SERVICE_NAME original value in case a test changes it
        -- need to use _G otherwise calling SERVICE_NAME here doesn't refer to 
        -- the variable defined in filter.lua
        _service = _G["SERVICE_NAME"]
    end)

    before_each(function()
        _G["SERVICE_NAME"] = "foo_service"
    end)

    after_each(function()
        -- restore service value if changed in test
        _G["SERVICE_NAME"] = _service
    end)

    it("returns if no contract specified", function()
        local reqHandler = NewRequestHandlerMock({
            request_headers = {
                host = "foo_host"
            }
        })
        spy.on(reqHandler, "httpCall")
        envoy_on_request(reqHandler)
        assert.spy(reqHandler.httpCall).was.not_called()
    end)

    it("returns if contract doesn't match this service name", function() 
        _G["SERVICE_NAME"] = "bar_service"
        local reqHandler = NewRequestHandlerMock({
            request_headers = {
                host = "foo_host",
                ['x-devroute'] = '{"foo_service": "127.0.0.1:9000"}'
            }
        })
        spy.on(reqHandler, "httpCall")
        envoy_on_request(reqHandler)
        assert.spy(reqHandler.httpCall).was.not_called()
    end)

    it("adds matched service header", function()
        local reqHandler = NewRequestHandlerMock({
            request_headers = {
                host = "foo_host",
                ['x-devroute'] = '{"foo_service": "127.0.0.1:9000"}'
            }
        })
        envoy_on_request(reqHandler)
        assert.are.equal("foo_service", reqHandler:headers():get("x-devroute-matched"))
    end)

    it("sends all original request's headers", function()
        local reqHeaders = {
            host = "foo_host",
            foo = "bar",
            qux = "42",
            ['x-devroute'] = '{"foo_service": "127.0.0.1:9000"}'
        }
        local reqHandler = NewRequestHandlerMock({
            request_headers = reqHeaders
        })
        spy.on(reqHandler, "httpCall")
        envoy_on_request(reqHandler)
        
        -- need to define custom matcher to check header equality
        -- see Extending matchers on https://olivinelabs.com/busted/#matchers
        local function sameHeaders(state, arguments) 
            local expected = arguments[1]
            return function(value)
                for k, v in pairs(expected) do
                    if value.hvalues[k] == nil then
                        return false
                    end
                end
                return true
            end
        end
        assert:register("matcher", "sameHeaders", sameHeaders)

        assert.spy(reqHandler.httpCall).was_called_with(_, _, match.sameHeaders(reqHeaders), _, _)
    end)

    it("sends original request's body", function()
        local body = "hello world"
        local reqHandler = NewRequestHandlerMock({
            request_headers = {
                host = "foo_host",
                ['x-devroute'] = '{"foo_service": "127.0.0.1:9000"}'
            },
            request_body = body
        })
        spy.on(reqHandler, "httpCall")
        envoy_on_request(reqHandler)

        assert.spy(reqHandler.httpCall).was_called_with(_, _, _, body, _)
    end)

    it("responds back with upstream's headers and body", function()
        local responseHeaders = {
            foo = "bar",
            lorem = "ipsum"
        }
        local responseBody = "Hello from upstream"
        local reqHandler = NewRequestHandlerMock({
            request_headers = {
                host = "foo_host",
                ['x-devroute'] = '{"foo_service": "127.0.0.1:9000"}'
            },
            response_body = responseBody,
            response_headers = responseHeaders
        })
        spy.on(reqHandler, "respond")
        envoy_on_request(reqHandler)

        assert.spy(reqHandler.respond).was_called()
        assert.spy(reqHandler.respond).was_called_with(_, responseHeaders, responseBody)
    end)
  end)

describe("get_req_body", function()
    it("returns empty string if no body in request", function()
        local reqHandler = NewRequestHandlerMock()
        assert.are.equal("", get_req_body(reqHandler))
    end)

    it("returns request body", function() 
        local body = "foo"
        local reqHandler = NewRequestHandlerMock({
            request_body = body
        })
        assert.are.equal(body, get_req_body(reqHandler))
    end)
end)

describe("escape_lua_magic", function()
    it("escapes magic characters", function()
        assert.are.equal("foo%-service%-bar", escape_lua_magic("foo-service-bar"))
    end)
end)
