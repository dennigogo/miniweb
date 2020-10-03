package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/dennigogo/miniweb/general"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	def    = `/`
	version  = `version`
	full   = `full`
	single = `single`
)

type config struct {
	Port     int    `env:"MINIWEB_PORT" envDefault:"8808"`
	LogLevel string `env:"MINIWEB_LOG_LEVEL" envDefault:"debug"`
	LogLines bool   `env:"MINIWEB_LOG_LINES" envDefault:"true"`
	LogJson  bool   `env:"MINIWEB_LOG_JSON" envDefault:"true"`
}

func main() {
	var err error

	var conf config
	if err := env.Parse(&conf); err != nil {
		panic(err)
	}

	logger := logrus.New()

	level, _ := logrus.ParseLevel(conf.LogLevel)
	if conf.LogJson {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}
	logger.SetLevel(level)
	logger.SetReportCaller(conf.LogLines)

	logs := logrus.NewEntry(logger).WithField(`module`, `miniweb`)

	router := mux.NewRouter()

	defsRoute := func(w http.ResponseWriter, r *http.Request) {
		log := logs.WithFields(setLogs(r))

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.WithError(err).Error(`io utils read all`)
			return
		}

		log.WithField(`raw`, fmt.Sprintf(`%+x`, r.Body)).Info(string(bodyBytes))

		var url string
 		if def != r.URL.String() {
			url = r.URL.String()
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`Hello %s%s`, r.Host, url)))
	}

	router.HandleFunc(def, defsRoute).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defsRoute(w, r)
	})

	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defsRoute(w, r)
	})

	router.HandleFunc(fmt.Sprintf(`/%s`, version),
		func(w http.ResponseWriter, r *http.Request) {
			log := logs.WithFields(setLogs(r))

			bytes, err := json.Marshal(&general.Versions{
				Version: general.Version,
				BuildTime: general.BuildTime,
			})
			if err != nil {
				logs.WithError(err).Error(`serialize data`)
				return
			}

			log.WithField(`raw`, fmt.Sprintf(`%+x`, r.Body)).Info(`get info`)

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
		}).Methods("GET")

	router.HandleFunc(fmt.Sprintf(`/%s`, full),
		func(w http.ResponseWriter, r *http.Request) {
			log := logs.WithFields(setLogs(r))

			defer func() {
				w.WriteHeader(http.StatusOK)
			}()

			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.WithError(err).Error(`io utils read all`)
				return
			}

			log.Info(string(bodyBytes))
		}).Methods("POST")

	router.HandleFunc(fmt.Sprintf(`/%s`, single),
		func(w http.ResponseWriter, r *http.Request) {
			log := logs.WithFields(setLogs(r))

			defer func() {
				w.WriteHeader(http.StatusOK)
			}()

			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.WithError(err).Error(`io utils read all`)
				return
			}

			log.Info(string(bodyBytes))
		}).Methods("POST")

	cancel, err := setHTTPServerByPort(conf.Port, router)
	if err != nil {
		logrus.WithField(`port`, conf.Port).WithError(err).Error(`set HTTP server by port`)
		cancel()
	}

	check := make(chan os.Signal)
	signal.Notify(check, syscall.SIGTERM, syscall.SIGINT)
	<-check
	cancel()
}

func setHTTPServerByPort(port int, router *mux.Router) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())

	server := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: router}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			return
		}

		if err := server.Shutdown(ctx); err != nil {
			return
		}
	}()

	return cancel, nil
}

func setLogs(r *http.Request) logrus.Fields {
	var url string
	if def != r.URL.String() {
		url = r.URL.String()
	}

	return logrus.Fields{
		`route`: fmt.Sprintf(`%s%s`, r.Host, url),
		`method`: r.Method,
	}
}