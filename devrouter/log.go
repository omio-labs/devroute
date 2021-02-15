package main

import (
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

// InitLogger configures log to output on os.Stdout and with
// json format
func InitLogger() {
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyTime: "timestamp",
			log.FieldKeyMsg:  "message",
		},
	})
	log.SetOutput(os.Stdout)
}

// LogRequestFields returns a log.Fields with
// useful information taken from r.
func LogRequestFields(r *http.Request) log.Fields {
	return log.Fields{
		"method":          r.Method,
		"URI":             r.RequestURI,
		"upstream":        r.Host,
		"contract":        r.Header.Get("X-Devroute"),
		"matched-service": r.Header.Get("X-Devroute-Matched"),
	}
}
