package conf

import (
	"flag"
	"fmt"
	"github.com/tietang/props/ini"
	"github.com/wcaqrl/chime/pkg/pather"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bilibili/discovery/naming"
	xcommon "github.com/wcaqrl/chime/pkg/common"
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
	debug     bool
	// Conf config
	Conf *Config
)

func __init() {
	var (
		defHost, _  = os.Hostname()
		defDebug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	)
	flag.BoolVar(&help, "h", false, "this help")
	flag.StringVar(&confPath, "c", "app.ini", "configuration file, default app.ini")
	flag.StringVar(&region, "r", os.Getenv("REGION"), "avaliable region. or use REGION env variable, value: sh etc.")
	flag.StringVar(&zone, "z", os.Getenv("ZONE"), "avaliable zone. or use ZONE env variable, value: sh001/sh002 etc.")
	flag.StringVar(&deployEnv, "e", os.Getenv("DEPLOY_ENV"), "deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	flag.StringVar(&host, "d", defHost, "machine hostname. or use default machine hostname.")
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
		Env:       &Env{Region: region, Zone: zone, DeployEnv: deployEnv, Host: host},
		Kafka:     &Kafka{Topic: "chime-push-topic", Group: "chime-push-group-job", Brokers: []string{}},
		Discovery: &naming.Config{Region: region, Zone: zone, Env: deployEnv, Host: host},
		Comet:     &Comet{RoutineChan: 1024, RoutineSize: 32},
		Room: &Room{
			Batch:  20,
			Signal: xtime.Duration(time.Second),
			Idle:   xtime.Duration(time.Minute * 15),
		},
	}
}

func initConfig() {
	var (
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
	// kafka
	Conf.Kafka.Topic = conf.GetDefault("kafka.topic", "chime-push-topic")
	Conf.Kafka.Group = conf.GetDefault("kafka.group", "chime-push-group-job")
	tmpStr = conf.GetDefault("kafka.brokers", "")
	if tmpStr != "" {
		Conf.Kafka.Brokers = strings.Split(tmpStr, ",")
	}
}

func usage() {
	var consoleStr = `
	chime-job version: 1.0.0
	Usage: chime-job [-h | -c | -r | -z | -e | -d | -b]
	   e.g.  : ./chime-job -c=app.ini -r=wh -z=wh01 -e=dev

	   -h    : this help
	   -c    : configuration file, default app.ini
	   -r    : avaliable region. or use REGION env variable, value: sh etc.
	   -z    : avaliable zone. or use ZONE env variable, value: sh001/sh002 etc.
	   -e    : deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.
	   -d    : machine hostname. or use default machine hostname.
	   -b    : server debug. or use DEBUG env variable, value: true/false etc.
`
	fmt.Fprintf(os.Stdout, consoleStr)
}

// Config is job config.
type Config struct {
	Debug     bool
	Logger    *xcommon.Logger
	Env       *Env
	Kafka     *Kafka
	Discovery *naming.Config
	Comet     *Comet
	Room      *Room
}

// Room is room config.
type Room struct {
	Batch  int
	Signal xtime.Duration
	Idle   xtime.Duration
}

// Comet is comet config.
type Comet struct {
	RoutineChan int
	RoutineSize int
}

// Kafka is kafka config.
type Kafka struct {
	Topic   string
	Group   string
	Brokers []string
}

// Env is env config.
type Env struct {
	Region    string
	Zone      string
	DeployEnv string
	Host      string
}
