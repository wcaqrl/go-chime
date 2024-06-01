package main

import (
	"context"
	"fmt"
	"github.com/bilibili/discovery/naming"
	resolver "github.com/bilibili/discovery/naming/grpc"
	log "github.com/sirupsen/logrus"
	"github.com/wcaqrl/chime/internal/comet"
	"github.com/wcaqrl/chime/internal/comet/conf"
	"github.com/wcaqrl/chime/internal/comet/grpc"
	md "github.com/wcaqrl/chime/internal/logic/model"
	"github.com/wcaqrl/chime/pkg/ip"
	"github.com/wcaqrl/chime/pkg/logger"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	ver   = "1.0.0"
	appid = "chime.comet"
)

func main() {
	if err := conf.Init(); err != nil {
		panic(err)
	}
	conf.Conf.Logger.AppID = appid
	logger.InitLogger(conf.Conf.Logger)
	rand.Seed(time.Now().UTC().UnixNano())
	var numCpu = runtime.GOMAXPROCS(runtime.NumCPU())
	action := "running"
	if conf.Conf.Debug {
		action = "debugging"
	}
	log.Infof("You are %s [%s %s] on %d cpus, env: %+v", action, appid, ver, numCpu, conf.Conf.Env)

	// register discovery
	dis := naming.New(conf.Conf.Discovery)
	resolver.Register(dis)
	// new comet server
	srv := comet.NewServer(conf.Conf)
	if err := comet.InitWhitelist(conf.Conf.Whitelist); err != nil {
		panic(err)
	}
	if err := comet.InitTCP(srv, conf.Conf.TCP.Bind, runtime.NumCPU()); err != nil {
		panic(err)
	}
	if err := comet.InitWebsocket(srv, conf.Conf.Websocket.Bind, runtime.NumCPU()); err != nil {
		panic(err)
	}
	if conf.Conf.Websocket.TLSOpen {
		if err := comet.InitWebsocketWithTLS(srv, conf.Conf.Websocket.TLSBind, conf.Conf.Websocket.CertFile, conf.Conf.Websocket.PrivateFile, runtime.NumCPU()); err != nil {
			panic(err)
		}
	}
	// new grpc server
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)
	cancel := register(dis, srv)
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("chime-comet get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if cancel != nil {
				cancel()
			}
			rpcSrv.GracefulStop()
			srv.Close()
			log.Infof("chime-comet [version: %s] exit", ver)
			// log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func register(dis *naming.Discovery, srv *comet.Server) context.CancelFunc {
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
			md.MetaWeight:  strconv.FormatInt(env.Weight, 10),
			md.MetaOffline: strconv.FormatBool(env.Offline),
			md.MetaAddrs:   strings.Join(env.Addrs, ","),
		},
	}
	cancel, err := dis.Register(ins)
	if err != nil {
		panic(err)
	}
	// renew discovery metadata
	go func() {
		for {
			var (
				wrong error
				conns int
				ips   = make(map[string]struct{})
			)
			for _, bucket := range srv.Buckets() {
				for tmpIp := range bucket.IPCount() {
					ips[tmpIp] = struct{}{}
				}
				conns += bucket.ChannelCount()
			}
			ins.Metadata[md.MetaConnCount] = fmt.Sprint(conns)
			ins.Metadata[md.MetaIPCount] = fmt.Sprint(len(ips))
			if wrong = dis.Set(ins); wrong != nil {
				log.Errorf("dis.Set(%+v) error(%v)", ins, wrong)
				time.Sleep(time.Second)
				continue
			}
			time.Sleep(time.Second * 10)
		}
	}()
	return cancel
}
