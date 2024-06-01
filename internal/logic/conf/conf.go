package conf

import (
	"flag"
	"fmt"
	"github.com/tietang/props/ini"
	xcommon "github.com/wcaqrl/chime/pkg/common"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bilibili/discovery/naming"
	"github.com/wcaqrl/chime/pkg/pather"
	xtime "github.com/wcaqrl/chime/pkg/time"
)

var (
	help      bool
	ePath     string
	confPath  string
	region    string
	zone      string
	deployEnv string
	host      string
	weight    int64
	debug     bool

	// Conf config
	Conf *Config
)

func __init() {
	var (
		defHost, _   = os.Hostname()
		defWeight, _ = strconv.ParseInt(os.Getenv("WEIGHT"), 10, 32)
		defDebug, _  = strconv.ParseBool(os.Getenv("DEBUG"))
	)
	flag.BoolVar(&help, "help", false, "this help")
	flag.StringVar(&confPath, "c", "app.ini", "default config path")
	flag.StringVar(&region, "r", os.Getenv("REGION"), "avaliable region. or use REGION env variable, value: sh etc.")
	flag.StringVar(&zone, "z", os.Getenv("ZONE"), "avaliable zone. or use ZONE env variable, value: sh001/sh002 etc.")
	flag.StringVar(&deployEnv, "e", os.Getenv("DEPLOY_ENV"), "deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	flag.StringVar(&host, "d", defHost, "machine hostname. or use default machine hostname.")
	flag.Int64Var(&weight, "w", defWeight, "load balancing weight, or use WEIGHT env variable, value: 10 etc.")
	flag.BoolVar(&debug, "b", defDebug, "server debug. or use DEBUG env variable, value: true/false etc.")
	flag.Parse()
}

// Init init config.
func Init() (err error) {
	__init()
	if help {
		flag.Usage = usage
		flag.Usage()
		os.Exit(0)
	}
	Conf = Default()
	if ePath, err = os.Getwd(); err != nil {
		panic(err)
	}
	if confPath, err = pather.GetConfigFile(confPath, ePath); err != nil {
		panic(err)
	}
	initConfig()
	return
}

func initConfig() {
	var (
		err    error
		tmpStr string
		conf   *ini.IniFileConfigSource
	)
	conf = ini.NewIniFileConfigSource(confPath)
	// logger
	Conf.Logger.Level = conf.GetDefault("log.level", "info")
	Conf.Logger.Path = pather.GetLogPath(ePath, conf.GetDefault("log.path", "./logs"))
	Conf.Logger.Save = conf.GetIntDefault("log.save", 7)
	// discovery
	tmpStr = conf.GetDefault("discovery.nodes", "")
	if tmpStr != "" {
		Conf.Discovery.Nodes = strings.Split(tmpStr, ",")
	}
	// rpc client
	tmpStr = conf.GetDefault("rpc_client.dial", "1s")
	if Conf.RPCClient.Dial, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.RPCClient.Dial = xtime.Duration(1e9)
	}
	tmpStr = conf.GetDefault("rpc_client.timeout", "1s")
	if Conf.RPCClient.Timeout, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.RPCClient.Timeout = xtime.Duration(1e9)
	}
	// rpc server
	Conf.RPCServer.Network = conf.GetDefault("rpc_server.network", "tcp")
	Conf.RPCServer.Addr = conf.GetDefault("rpc_server.addr", ":3119")
	tmpStr = conf.GetDefault("rpc_server.timeout", "1s")
	if Conf.RPCServer.Timeout, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.RPCServer.Timeout = xtime.Duration(1e9)
	}
	// http server
	Conf.HTTPServer.Network = conf.GetDefault("http_server.network", "tcp")
	Conf.HTTPServer.Addr = conf.GetDefault("http_server.addr", ":3111")
	tmpStr = conf.GetDefault("http_server.read_timeout", "1s")
	if Conf.HTTPServer.ReadTimeout, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.HTTPServer.ReadTimeout = xtime.Duration(1e9)
	}
	tmpStr = conf.GetDefault("http_server.write_timeout", "1s")
	if Conf.HTTPServer.WriteTimeout, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.HTTPServer.WriteTimeout = xtime.Duration(1e9)
	}
	// kafka
	Conf.Kafka.Topic = conf.GetDefault("kafka.topic", "chime-push-topic")
	tmpStr = conf.GetDefault("kafka.brokers", "")
	if tmpStr != "" {
		Conf.Kafka.Brokers = strings.Split(tmpStr, ",")
	}
	// redis
	Conf.Redis.Network = conf.GetDefault("redis.network", "tcp")
	Conf.Redis.Addr = conf.GetDefault("redis.addr", ":6379")
	Conf.Redis.Db = conf.GetIntDefault("redis.db", 0)
	Conf.Redis.Active = conf.GetIntDefault("redis.active", 6000)
	Conf.Redis.Idle = conf.GetIntDefault("redis.idle", 1024)
	tmpStr = conf.GetDefault("redis.dial_timeout", "200ms")
	if Conf.Redis.DialTimeout, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.Redis.DialTimeout = xtime.Duration(200 * 1e6)
	}
	tmpStr = conf.GetDefault("redis.read_timeout", "500ms")
	if Conf.Redis.ReadTimeout, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.Redis.ReadTimeout = xtime.Duration(500 * 1e6)
	}
	tmpStr = conf.GetDefault("redis.write_timeout", "500ms")
	if Conf.Redis.WriteTimeout, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.Redis.WriteTimeout = xtime.Duration(500 * 1e6)
	}
	tmpStr = conf.GetDefault("redis.idle_timeout", "120s")
	if Conf.Redis.IdleTimeout, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.Redis.IdleTimeout = xtime.Duration(120 * 1e9)
	}
	tmpStr = conf.GetDefault("redis.expire", "30m")
	if Conf.Redis.Expire, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.Redis.Expire = xtime.Duration(30 * 60 * 1e9)
	}
	// node
	Conf.Node.DefaultDomain = conf.GetDefault("node.default_domain", "conn.chime.io")
	Conf.Node.HostDomain = conf.GetDefault("node.host_domain", ".chime.io")
	Conf.Node.TCPPort = conf.GetIntDefault("node.tcp_port", 3101)
	Conf.Node.WSPort = conf.GetIntDefault("node.ws_port", 3102)
	Conf.Node.WSSPort = conf.GetIntDefault("node.wss_port", 3103)
	Conf.Node.HeartbeatMax = conf.GetIntDefault("node.heartbeat_max", 2)
	tmpStr = conf.GetDefault("node.heartbeat", "4m")
	if Conf.Node.Heartbeat, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.Node.Heartbeat = xtime.Duration(4 * 60 * 1e9)
	}
	Conf.Node.RegionWeight = conf.GetFloat64Default("node.region_weight", 1.6)
	// backoff
	Conf.Backoff.MaxDelay = int32(conf.GetIntDefault("backoff.max_delay", 300))
	Conf.Backoff.BaseDelay = int32(conf.GetIntDefault("backoff.base_delay", 3))
	Conf.Backoff.Factor = float32(conf.GetFloat64Default("backoff.factor", 1.8))
	Conf.Backoff.Jitter = float32(conf.GetFloat64Default("backoff.jitter", 1.8))
	// regions
	Conf.Regions = parseRegions(conf)
}

func usage() {
	var consoleStr = `
	chime-job version: 1.0.0
	Usage: chime-job [-h | -c | -r | -z | -e | -d | -w | -b]
	   e.g.  : ./chime-job -c=app.ini -r=wh -z=wh01 -e=dev -w=10

	   -h    : this help
	   -c    : configuration file, default app.ini
	   -r    : avaliable region. or use REGION env variable, value: sh etc.
	   -z    : avaliable zone. or use ZONE env variable, value: sh001/sh002 etc.
	   -e    : deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.
	   -d    : machine hostname. or use default machine hostname.
	   -w    : load balancing weight, or use WEIGHT env variable, value: 10 etc.
	   -b    : server debug. or use DEBUG env variable, value: true/false etc.
`
	fmt.Fprintf(os.Stdout, consoleStr)
}

// Default new a config with specified default value.
func Default() *Config {
	return &Config{
		Debug:     debug,
		Logger:    &xcommon.Logger{Level: "info", Path: "./logs", Save: 7},
		Env:       &Env{Region: region, Zone: zone, DeployEnv: deployEnv, Host: host, Weight: weight},
		Discovery: &naming.Config{Region: region, Zone: zone, Env: deployEnv, Host: host},
		HTTPServer: &HTTPServer{
			Network:      "tcp",
			Addr:         "3111",
			ReadTimeout:  xtime.Duration(time.Second),
			WriteTimeout: xtime.Duration(time.Second),
		},
		RPCClient: &RPCClient{Dial: xtime.Duration(time.Second), Timeout: xtime.Duration(time.Second)},
		RPCServer: &RPCServer{
			Network:           "tcp",
			Addr:              "3119",
			Timeout:           xtime.Duration(time.Second),
			IdleTimeout:       xtime.Duration(time.Second * 60),
			MaxLifeTime:       xtime.Duration(time.Hour * 2),
			ForceCloseWait:    xtime.Duration(time.Second * 20),
			KeepAliveInterval: xtime.Duration(time.Second * 60),
			KeepAliveTimeout:  xtime.Duration(time.Second * 20),
		},
		Kafka:   &Kafka{},
		Redis:   &Redis{},
		Node:    &Node{},
		Backoff: &Backoff{MaxDelay: 300, BaseDelay: 3, Factor: 1.8, Jitter: 1.3},
		Regions: map[string][]string{},
	}
}

func parseRegions(conf *ini.IniFileConfigSource) (regions map[string][]string) {
	regions = make(map[string][]string)
	for _, key := range conf.Keys() {
		if strings.HasPrefix(key, "regions") {
			strArr := strings.Split(key, ".")
			regions[strArr[1]] = strings.Split(conf.GetDefault(key, ""), ",")
		}
	}
	return
}

// Config config.
type Config struct {
	Debug      bool
	Logger     *xcommon.Logger
	Env        *Env
	Discovery  *naming.Config
	RPCClient  *RPCClient
	RPCServer  *RPCServer
	HTTPServer *HTTPServer
	Kafka      *Kafka
	Redis      *Redis
	Node       *Node
	Backoff    *Backoff
	Regions    map[string][]string
}

// Env is env config.
type Env struct {
	Region    string
	Zone      string
	DeployEnv string
	Host      string
	Weight    int64
}

// Node node config.
type Node struct {
	DefaultDomain string
	HostDomain    string
	TCPPort       int
	WSPort        int
	WSSPort       int
	HeartbeatMax  int
	Heartbeat     xtime.Duration
	RegionWeight  float64
}

// Backoff backoff.
type Backoff struct {
	MaxDelay  int32
	BaseDelay int32
	Factor    float32
	Jitter    float32
}

// Redis .
type Redis struct {
	Network      string
	Addr         string
	Db           int
	Auth         string
	Active       int
	Idle         int
	DialTimeout  xtime.Duration
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
	IdleTimeout  xtime.Duration
	Expire       xtime.Duration
}

// Kafka .
type Kafka struct {
	Topic   string
	Brokers []string
}

// RPCClient is RPC client config.
type RPCClient struct {
	Dial    xtime.Duration
	Timeout xtime.Duration
}

// RPCServer is RPC server config.
type RPCServer struct {
	Network           string
	Addr              string
	Timeout           xtime.Duration
	IdleTimeout       xtime.Duration
	MaxLifeTime       xtime.Duration
	ForceCloseWait    xtime.Duration
	KeepAliveInterval xtime.Duration
	KeepAliveTimeout  xtime.Duration
}

// HTTPServer is http server config.
type HTTPServer struct {
	Network      string
	Addr         string
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
}
