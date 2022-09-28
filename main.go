package main

import (
	"bbs/pkg/cache"
	"bbs/pkg/global"
	"bbs/pkg/jwt"
	"bbs/pkg/logging"
	"bbs/pkg/mysql"
	"bbs/router"
	"context"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	global.LoadConfig()

	global.LOG = logging.SetupLogger()

	logging.Init()

	//初始化redis
	err := cache.InitRedis(cache.DefaultRedisClient, &redis.Options{
		Addr:        global.CONFIG.Redis.Host,
		Password:    global.CONFIG.Redis.Password,
		IdleTimeout: global.CONFIG.Redis.IdleTimeout,
	}, nil)
	if err != nil {
		global.LOG.Error("InitRedis error", err, "client", cache.DefaultRedisClient)
		panic(err)
	}

	//初始化mysql
	err = mysql.InitMysqlClient(mysql.DefaultClient, global.CONFIG.Database.User,
		global.CONFIG.Database.Password, global.CONFIG.Database.Host,
		global.CONFIG.Database.Name)
	if err != nil {
		global.LOG.Error("InitMysqlClient error", err, "client", mysql.DefaultClient)
	}
	global.Db = mysql.GetMysqlClient(mysql.DefaultClient).DB

	jwt.Init()

}

func main() {
	gin.SetMode(global.CONFIG.Server.RunMode)

	routersInit := router.InitRouter()
	endPoint := fmt.Sprintf(":%d", global.CONFIG.Server.HttpPort)
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           endPoint,
		Handler:        routersInit,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logging.Error("start http server error", err)
		} else {
			fmt.Println("start http server listening", endPoint)
		}
	}()

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	zap.L().Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown:")
	}
	zap.L().Info("Server exiting")
	// 6. 启动服务

	//优雅关闭
	//shutdown.NewHook().Close(
	//	//关闭http server
	//	func() {
	//		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	//		defer cancel()
	//		if err := server.Shutdown(ctx); err != nil {
	//			logging.Error("http server shutdown error", err)
	//		}
	//	},
	//	//关闭kafka producer
	//	func() {
	//		if err := mq.GetKafkaSyncProducer(mq.DefaultKafkaSyncProducer).Close(); err != nil {
	//			logging.Error("kafka close error", err, "client", mq.DefaultKafkaSyncProducer)
	//		}
	//	},
	//	//关闭mysql
	//	func() {
	//		if err := db.CloseMysqlClient(db.DefaultClient); err != nil {
	//			logging.Error("CloseMysqlClient error", err, "client", db.DefaultClient)
	//		}
	//	},
	//)

}
