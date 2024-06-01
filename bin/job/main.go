package main

import (
	"github.com/wcaqrl/chime/pkg/logger"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/bilibili/discovery/naming"
	"github.com/wcaqrl/chime/internal/job"
	"github.com/wcaqrl/chime/internal/job/conf"

	resolver "github.com/bilibili/discovery/naming/grpc"
	log "github.com/sirupsen/logrus"
)

const (
	ver   = "1.0.0"
	appid = "chime.job"
)

func main() {
	if err := conf.Init(); err != nil {
		panic(err)
	}
	conf.Conf.Logger.AppID = appid
	logger.InitLogger(conf.Conf.Logger)
	var numCpu = runtime.GOMAXPROCS(runtime.NumCPU())
	action := "running"
	if conf.Conf.Debug {
		action = "debugging"
	}
	log.Infof("You are %s [%s %s] on %d cpus, env: %+v", action, appid, ver, numCpu, conf.Conf.Env)

	// grpc register naming
	dis := naming.New(conf.Conf.Discovery)
	resolver.Register(dis)
	// job
	j := job.New(conf.Conf)
	go j.Consume()
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("chime-job get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			j.Close()
			log.Infof("chime-job [version: %s] exit", ver)
			// log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
