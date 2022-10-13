package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"lovebox/models"
	"lovebox/pkg/gin/middlewares"
	"lovebox/pkg/validator"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	_ "google.golang.org/grpc/encoding/gzip" // Install the gzip compressor
)

type serverFlags struct {
	port string
}

var sFlags serverFlags

func init() {
	serverCmd.Flags().StringVar(&sFlags.port, "port", "8008", "listen port")
	_ = viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))

	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:     "api",
	Aliases: []string{"api"},
	Short:   "Start api application",
	Run: func(cmd *cobra.Command, args []string) {
		log := zap.S().With("cmd", "api")
		pkgs := NewPackages()
		svcs := NewServices(pkgs)
		ginCtrls := NewGinControllers(pkgs, svcs)

		// 创建mysql数据库表
		if err := pkgs.mysqlClient.AutoMigrate(
			&models.Account{},
			&models.AccountExtraInfo{},
			&models.OperateLogs{},
		); err != nil {
			log.Fatalf("Mysql AutoMigrate Error: %v", err)
		}

		var httpPublicServer *http.Server

		var eg errgroup.Group
		eg.Go(func() error {
			engine, err := GetGinPublicEngine(ginCtrls, pkgs)
			if err != nil {
				return err
			}

			port := ":" + viper.GetString("http.public.port")
			log.Infof("HTTP public server listen on %s", port)
			httpPublicServer = &http.Server{
				Addr:    port,
				Handler: engine,
			}

			return httpPublicServer.ListenAndServe()
		})

		eg.Go(func() error {
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
			sig := <-signalChan

			log.Warnf("System about to exit because of signal=%s", sig.String())

			var wg sync.WaitGroup
			servers := []*http.Server{
				httpPublicServer,
			}

			for _, svr := range servers {
				if svr == nil {
					continue
				}

				svr := svr
				wg.Add(1)

				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if err := svr.Shutdown(ctx); err != nil {
						log.Errorf("Api server shutdown err=%v", err)
					}
					wg.Done()
				}()
			}

			wg.Wait()

			return nil
		})

		if err := eg.Wait(); err != nil {
			log.Fatalf("Server err %v", err)
		}
	},
}

func GetGinPublicEngine(ctrls *GinControllers, pkgs *Packages) (*gin.Engine, error) {
	if os.Getenv("GO_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	_ = router.SetTrustedProxies(viper.GetStringSlice("http.trustedProxies"))

	router.Use(gin.Recovery())
	router.Use(cors.AllowAll())
	router.Use(middlewares.HTTPGzipEncoding)

	api := router.Group("/api/v1")
	pkgs.prom.Use(router)
	api.Use(pkgs.prom.Instrument("public"))
	pprof.RouteRegister(api)

	gzipExcludePaths := []string{}
	api.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths(gzipExcludePaths)))

	uni, err := validator.GetUniversalTranslator()
	if err != nil {
		return nil, err
	}
	eh := middlewares.NewWithStatusHandler(uni)
	api.Use(eh.HandleErrors)

	api.Use(middlewares.Logger(zap.S()))
	api.Use(middlewares.NewPaginationMiddleware())
	api.Use(middlewares.NewI18nMiddleware())
	api.Use(middlewares.Tracing(middlewares.TracingComponentName("gin")))
	api.Use(middlewares.NewOperateLogger(zap.S(), pkgs.mysqlClient))

	api.GET("captcha", ctrls.accountCtrl.GetCaptcha)
	api.POST("login", ctrls.accountCtrl.Login)
	api.POST("register", ctrls.accountCtrl.Register)
	api.Use(middlewares.NewJwtCheckMiddleware(pkgs.jwt, pkgs.mysqlClient, pkgs.cacheClient))
	api.GET("info", ctrls.accountCtrl.Info)

	// base := api.Group("")

	return router, nil
}
