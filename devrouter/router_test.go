package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestProxyToDev(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	proxy := httptest.NewServer(http.HandlerFunc(ProxyToDev))
	defer proxy.Close()
	proxyURL, _ := url.Parse(proxy.URL)

	t.Run("ItRejectsMissingContract", func(t *testing.T) {
		resp, _ := http.Get(proxyURL.String())
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ItRejectsMalformedContract", func(t *testing.T) {
		req := &http.Request{
			Method: "GET",
			Header: map[string][]string{"X-Devroute": {`{"foo": [1,2,3]}`}},
			URL:    proxyURL,
		}
		resp, _ := http.DefaultClient.Do(req)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ItRejectsMissingMatchedServiceInContract", func(t *testing.T) {
		req := &http.Request{
			Method: "GET",
			Header: map[string][]string{
				"X-Devroute":         {`{"foo": "192.168.10.20:9001"}`},
				"X-Devroute-Matched": {"bar"}},
			URL: proxyURL,
		}
		resp, _ := http.DefaultClient.Do(req)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ItRejectsMalformedIPAndPort", func(t *testing.T) {
		req := &http.Request{
			Method: "GET",
			Header: map[string][]string{
				"X-Devroute":         {`{"foo": "192.168.10.20_9001"}`},
				"X-Devroute-Matched": {"foo"}},
			URL: proxyURL,
		}
		resp, _ := http.DefaultClient.Do(req)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ItRejectsRequestsWithNonPrivateIP", func(t *testing.T) {
		req := &http.Request{
			Method: "GET",
			Header: map[string][]string{
				"X-Devroute":         {`{"foo": "69.145.32.56:9001"}`},
				"X-Devroute-Matched": {"foo"}},
			URL: proxyURL,
		}
		resp, _ := http.DefaultClient.Do(req)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusForbidden, resp.StatusCode)
		}
	})

	t.Run("ItProxiesToMatchedServiceInContract=1", func(t *testing.T) {
		serviceBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "pong")
		}))
		defer serviceBackend.Close()
		serviceURL := strings.TrimPrefix(serviceBackend.URL, "http://")

		proxyURLWithPath, _ := url.Parse(proxy.URL + "/ping")
		req := &http.Request{
			Method: "GET",
			Header: map[string][]string{
				"X-Devroute":         {fmt.Sprintf(`{"foo": "%s"}`, serviceURL)},
				"X-Devroute-Matched": {"foo"}},
			URL: proxyURLWithPath,
		}
		resp, _ := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if string(body) != "pong" {
			t.Errorf("Expected pong, got %s", string(body))
		}
	})

	// When having more than one component in contract
	t.Run("ItProxiesToMatchedServiceInContract=2", func(t *testing.T) {
		serviceBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "pong")
		}))
		serviceURL := strings.TrimPrefix(serviceBackend.URL, "http://")
		defer serviceBackend.Close()

		proxyURLWithPath, _ := url.Parse(proxy.URL + "/ping")
		req := &http.Request{
			Method: "GET",
			Header: map[string][]string{
				"X-Devroute":         {fmt.Sprintf(`{"foo": "%s", "bar":"127.0.0.1:4001"}`, serviceURL)},
				"X-Devroute-Matched": {"foo"}},
			URL: proxyURLWithPath,
		}
		resp, _ := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if string(body) != "pong" {
			t.Errorf("Expected pong, got %s", string(body))
		}
	})
}
