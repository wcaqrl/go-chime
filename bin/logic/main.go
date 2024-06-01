package main

import (
	"context"
	"github.com/bilibili/discovery/naming"
	resolver "github.com/bilibili/discovery/naming/grpc"
	log "github.com/sirupsen/logrus"
	"github.com/wcaqrl/chime/internal/logic"
	"github.com/wcaqrl/chime/internal/logic/conf"
	"github.com/wcaqrl/chime/internal/logic/grpc"
	"github.com/wcaqrl/chime/internal/logic/http"
	"github.com/wcaqrl/chime/internal/logic/model"
	"github.com/wcaqrl/chime/pkg/ip"
	"github.com/wcaqrl/chime/pkg/logger"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
)

const (
	ver   = "1.0.0"
	appid = "chime.logic"
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
	// logic
	srv := logic.New(conf.Conf)
	httpSrv := http.New(conf.Conf.HTTPServer, srv)
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)
	cancel := register(dis, srv)
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("chime-logic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if cancel != nil {
				cancel()
			}
			srv.Close()
			httpSrv.Close()
			rpcSrv.GracefulStop()
			log.Infof("chime-logic [version: %s] exit", ver)
			// log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func register(dis *naming.Discovery, srv *logic.Logic) context.CancelFunc {
	env := conf.Conf.Env
	addr := ip.InternalIP()
	_, port, _ := net.SplitHostPort(conf.Conf.RPCServer.Addr)
	ins := &naming.Instance{
		Region:   env.Region,
		Zone:     env.Zone,
		Env:      env.DeployEnv,
		Hostname: env.Host,
		AppID:    appid,
		Addrs: []string{
			"grpc://" + addr + ":" + port,
		},
		Metadata: map[string]string{
			model.MetaWeight: strconv.FormatInt(env.Weight, 10),
		},
	}
	cancel, err := dis.Register(ins)
	if err != nil {
		panic(err)
	}
	return cancel
}
