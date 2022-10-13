package cmd

import (
	"lovebox/pkg/database"
	"lovebox/pkg/gin/middlewares"
	"lovebox/pkg/jwt"
	"lovebox/pkg/tracing"
	"lovebox/pkg/utils"
	"lovebox/services/account"
	"lovebox/services/system"

	redisCache "github.com/go-redis/cache/v8"
	redislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

type Packages struct {
	mysqlClient   *database.Client
	redisClient   *redislib.Client
	cacheClient   *redisCache.Cache
	redSyncClient *redsync.Redsync
	prom          *middlewares.Prometheus
	tracing       *tracing.TracingService
	jwt           *jwt.Jwt
}

func NewPackages() (pkgs *Packages) {
	pkgs = &Packages{
		prom: middlewares.NewPromMiddleware(),
	}
	log := zap.S().With("module", "init")

	{
		pkgs.tracing = tracing.NewTracingService(
			viper.GetBool("tracing.ext.logging.enable"),
		)

		err := pkgs.tracing.InitGlobal(viper.Sub("tracing"))
		if err != nil {
			log.Errorf("Init tracing err %+v", err)
			panic(err)
		}
	}

	{
		viper.SetDefault("mysql.logLevel", logger.Info)
		mysqlClient, err := database.NewMysql(
			viper.GetString("mysql.addr"),
			viper.GetInt("mysql.maxIdleConns"),
			viper.GetInt("mysql.maxOpenConns"),
			viper.GetDuration("mysql.connMaxLifetime"),
			logger.LogLevel(viper.GetUint("mysql.logLevel")),
		)
		if err != nil {
			log.Errorf("Init mysql error %v", err)
			panic(err)
		}
		pkgs.mysqlClient = mysqlClient
	}

	{
		viper.SetDefault("jwt.key", "suanzi")
		viper.SetDefault("jwt.issue", "SZKJ")
		pkgs.jwt = jwt.New(
			[]byte(viper.GetString("jwt.key")),
			viper.GetString("jwt.issue"),
		)
	}

	{
		viper.SetDefault("redis.uri", "192.168.16.131:6379")
		viper.SetDefault("redis.password", "")
		viper.SetDefault("redis.db", 0)
		rdb := redislib.NewClient(&redislib.Options{
			Addr:     viper.GetString("redis.uri"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
		})
		pkgs.redisClient = rdb
		pkgs.cacheClient = redisCache.New(&redisCache.Options{
			Redis: rdb,
		})
		pkgs.redSyncClient = redsync.New(goredis.NewPool(pkgs.redisClient))
	}

	return
}

type Services struct {
	accountSvc *account.Service
	systemSvc  *system.Service
}

func NewServices(pkgs *Packages) *Services {
	systemSvc := system.NewService(
		pkgs.mysqlClient,
	)

	viper.SetDefault("login.rsaPrivateKey", utils.DefaultRSAPrivateKey)
	viper.SetDefault("login.rsaPublicKey", utils.DefaultRSAPublicKey)
	accountSvc := account.NewService(
		pkgs.mysqlClient,
		pkgs.redisClient,
		pkgs.cacheClient,
		pkgs.jwt,
	)

	return &Services{
		accountSvc: accountSvc,
		systemSvc:  systemSvc,
	}
}

type GinControllers struct {
	accountCtrl *account.GinController
	systemCtrl  *system.GinController
}

func NewGinControllers(pkgs *Packages, svcs *Services) *GinControllers {
	return &GinControllers{
		accountCtrl: account.NewGinController(
			svcs.accountSvc,
			svcs.systemSvc,
		),
		systemCtrl: system.NewGinController(
			svcs.systemSvc,
		),
	}
}
