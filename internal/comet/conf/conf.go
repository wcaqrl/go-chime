package conf

import (
	"flag"
	"fmt"
	"github.com/bilibili/discovery/naming"
	"github.com/tietang/props/ini"
	xcommon "github.com/wcaqrl/chime/pkg/common"
	"github.com/wcaqrl/chime/pkg/pather"
	xtime "github.com/wcaqrl/chime/pkg/time"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	help      bool
	ePath     string
	confPath  string
	region    string
	zone      string
	deployEnv string
	host      string
	addrs     string
	weight    int64
	offline   bool
	debug     bool

	// Conf config
	Conf *Config
)

func __init() {
	var (
		defHost, _    = os.Hostname()
		defAddrs      = os.Getenv("ADDRS")
		defWeight, _  = strconv.ParseInt(os.Getenv("WEIGHT"), 10, 32)
		defOffline, _ = strconv.ParseBool(os.Getenv("OFFLINE"))
		defDebug, _   = strconv.ParseBool(os.Getenv("DEBUG"))
	)
	flag.BoolVar(&help, "h", false, "this help")
	flag.StringVar(&confPath, "c", "app.ini", "configuration file, default app.ini")
	flag.StringVar(&region, "r", os.Getenv("REGION"), "avaliable region. or use REGION env variable, value: sh etc.")
	flag.StringVar(&zone, "z", os.Getenv("ZONE"), "avaliable zone. or use ZONE env variable, value: sh001/sh002 etc.")
	flag.StringVar(&deployEnv, "e", os.Getenv("DEPLOY_ENV"), "deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	flag.StringVar(&host, "d", defHost, "machine hostname. or use default machine hostname.")
	flag.StringVar(&addrs, "a", defAddrs, "server public ip addrs. or use ADDRS env variable, value: 127.0.0.1 etc.")
	flag.Int64Var(&weight, "w", defWeight, "load balancing weight, or use WEIGHT env variable, value: 10 etc.")
	flag.BoolVar(&offline, "o", defOffline, "server offline. or use OFFLINE env variable, value: true/false etc.")
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

// Default new a config with specified default value.
func Default() *Config {
	return &Config{
		Debug:     debug,
		Logger:    &xcommon.Logger{Level: "info", Path: "./logs", Save: 7},
		Env:       &Env{Region: region, Zone: zone, DeployEnv: deployEnv, Host: host, Weight: weight, Addrs: strings.Split(addrs, ","), Offline: offline},
		Discovery: &naming.Config{Region: region, Zone: zone, Env: deployEnv, Host: host},
		RPCClient: &RPCClient{
			Dial:    xtime.Duration(time.Second),
			Timeout: xtime.Duration(time.Second),
		},
		RPCServer: &RPCServer{
			Network:           "tcp",
			Addr:              ":3109",
			Timeout:           xtime.Duration(time.Second),
			IdleTimeout:       xtime.Duration(time.Second * 60),
			MaxLifeTime:       xtime.Duration(time.Hour * 2),
			ForceCloseWait:    xtime.Duration(time.Second * 20),
			KeepAliveInterval: xtime.Duration(time.Second * 60),
			KeepAliveTimeout:  xtime.Duration(time.Second * 20),
		},
		TCP: &TCP{
			Bind:         []string{":3101"},
			Sndbuf:       4096,
			Rcvbuf:       4096,
			KeepAlive:    false,
			Reader:       32,
			ReadBuf:      1024,
			ReadBufSize:  8192,
			Writer:       32,
			WriteBuf:     1024,
			WriteBufSize: 8192,
		},
		Websocket: &Websocket{
			Bind: []string{":3102"},
		},
		Protocol: &Protocol{
			Timer:            32,
			TimerSize:        2048,
			CliProto:         5,
			SvrProto:         10,
			HandshakeTimeout: xtime.Duration(time.Second * 5),
		},
		Bucket: &Bucket{
			Size:          32,
			Channel:       1024,
			Room:          1024,
			RoutineAmount: 32,
			RoutineSize:   1024,
		},
	}
}

func initConfig() {
	var (
		err         error
		tmpInt64    int64
		tmpStr      string
		tmpStrSlice []string
		conf        *ini.IniFileConfigSource
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
	tmpStr = conf.GetDefault("rpc_server.timeout", "1s")
	if Conf.RPCServer.Timeout, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.RPCServer.Timeout = xtime.Duration(1e9)
	}
	Conf.RPCServer.Addr = conf.GetDefault("rpc_server.addr", ":3109")
	// tcp
	tmpStr = conf.GetDefault("tcp.bind", ":3101")
	if tmpStr != "" {
		Conf.TCP.Bind = strings.Split(tmpStr, ",")
	}
	Conf.TCP.Sndbuf = conf.GetIntDefault("tcp.send_buffer", 4096)
	Conf.TCP.Rcvbuf = conf.GetIntDefault("tcp.receive_buffer", 4096)
	Conf.TCP.KeepAlive = conf.GetBoolDefault("tcp.keepalive", false)
	Conf.TCP.Reader = conf.GetIntDefault("tcp.reader", 32)
	Conf.TCP.ReadBuf = conf.GetIntDefault("tcp.read_buffer", 1024)
	Conf.TCP.ReadBufSize = conf.GetIntDefault("tcp.read_buffer_size", 8192)
	Conf.TCP.Writer = conf.GetIntDefault("tcp.writer", 32)
	Conf.TCP.WriteBuf = conf.GetIntDefault("tcp.write_buffer", 1024)
	Conf.TCP.WriteBufSize = conf.GetIntDefault("tcp.write_buffer_size", 8192)
	// websocket
	tmpStr = conf.GetDefault("websocket.bind", ":3102")
	if tmpStr != "" {
		Conf.Websocket.Bind = strings.Split(tmpStr, ",")
	}
	Conf.Websocket.TLSOpen = conf.GetBoolDefault("websocket.tls_open", false)
	tmpStr = conf.GetDefault("websocket.tls_bind", ":3103")
	if tmpStr != "" {
		Conf.Websocket.TLSBind = strings.Split(tmpStr, ",")
	}
	Conf.Websocket.CertFile = conf.GetDefault("websocket.cert_file", "../../cert.pem")
	Conf.Websocket.PrivateFile = conf.GetDefault("websocket.private_file", "../../private.pem")
	// protocol
	Conf.Protocol.Timer = conf.GetIntDefault("protocol.timer", 32)
	Conf.Protocol.TimerSize = conf.GetIntDefault("protocol.timer_size", 2048)
	Conf.Protocol.CliProto = conf.GetIntDefault("protocol.client_proto", 5)
	Conf.Protocol.SvrProto = conf.GetIntDefault("protocol.server_proto", 10)
	tmpStr = conf.GetDefault("protocol.handshake_timeout", "8s")
	if Conf.Protocol.HandshakeTimeout, err = xtime.UnmarshalDuration(tmpStr); err != nil {
		Conf.Protocol.HandshakeTimeout = xtime.Duration(8 * 1e9)
	}
	// bucket
	Conf.Bucket.Size = conf.GetIntDefault("bucket.size", 32)
	Conf.Bucket.Channel = conf.GetIntDefault("bucket.channel", 1024)
	Conf.Bucket.Room = conf.GetIntDefault("bucket.room", 1024)
	Conf.Bucket.RoutineAmount = uint64(conf.GetIntDefault("bucket.routine_amount", 32))
	Conf.Bucket.RoutineSize = conf.GetIntDefault("bucket.routine_size", 1024)
	// whitelist
	tmpStr = conf.GetDefault("whitelist.white_list", "")
	Conf.Whitelist = &Whitelist{
		Whitelist: make([]int64, 0),
		WhiteLog:  conf.GetDefault("whitelist.white_log", "/tmp/white_list.log"),
	}
	if tmpStr != "" {
		tmpStrSlice = strings.Split(tmpStr, ",")
		for _, v := range tmpStrSlice {
			if tmpInt64, err = strconv.ParseInt(v, 10, 64); err == nil {
				Conf.Whitelist.Whitelist = append(Conf.Whitelist.Whitelist, tmpInt64)
			}
		}
	}
}

func usage() {
	var consoleStr = `
	chime-comet version: 1.0.0
	Usage: chime-comet [-h | -c | -r | -z | -e | -d | -a | -w | -o | -b]
	   e.g.  : ./chime-comet -c=app.ini -r=wh -z=wh01 -e=dev -a=172.17.7.133 -w=10

	   -h    : this help
	   -c    : configuration file, default app.ini
	   -r    : avaliable region. or use REGION env variable, value: sh etc.
	   -z    : avaliable zone. or use ZONE env variable, value: sh001/sh002 etc.
	   -e    : deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.
	   -d    : machine hostname. or use default machine hostname.
	   -a    : server public ip addrs. or use ADDRS env variable, value: 127.0.0.1 etc.
	   -w    : load balancing weight, or use WEIGHT env variable, value: 10 etc.
	   -o    : server offline. or use OFFLINE env variable, value: true/false etc.
	   -b    : server debug. or use DEBUG env variable, value: true/false etc.
`
	fmt.Fprintf(os.Stdout, consoleStr)
}

// Config is comet config.
type Config struct {
	Debug     bool
	Logger    *xcommon.Logger
	Env       *Env
	Discovery *naming.Config
	TCP       *TCP
	Websocket *Websocket
	Protocol  *Protocol
	Bucket    *Bucket
	RPCClient *RPCClient
	RPCServer *RPCServer
	Whitelist *Whitelist
}

// Env is env config.
type Env struct {
	Region    string
	Zone      string
	DeployEnv string
	Host      string
	Weight    int64
	Offline   bool
	Addrs     []string
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

// TCP is tcp config.
type TCP struct {
	Bind         []string
	Sndbuf       int
	Rcvbuf       int
	KeepAlive    bool
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

// Websocket is websocket config.
type Websocket struct {
	Bind        []string
	TLSOpen     bool
	TLSBind     []string
	CertFile    string
	PrivateFile string
}

// Protocol is protocol config.
type Protocol struct {
	Timer            int
	TimerSize        int
	SvrProto         int
	CliProto         int
	HandshakeTimeout xtime.Duration
}

// Bucket is bucket config.
type Bucket struct {
	Size          int
	Channel       int
	Room          int
	RoutineAmount uint64
	RoutineSize   int
}

// Whitelist is white list config.
type Whitelist struct {
	Whitelist []int64
	WhiteLog  string
}
