package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"path"
	"strings"
	"system-conf/api"
	"system-conf/common/log"
)

type Args struct {
	Port int
}

func handleDocs(c *gin.Context) {
	if strings.HasSuffix(c.Request.RequestURI, "/api/docs/") || strings.HasSuffix(c.Request.RequestURI, "/api/docs") {
		c.Redirect(http.StatusTemporaryRedirect, path.Join(c.Request.RequestURI, "index.html"))
	} else {
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.InstanceName("systemconf"))(c)
	}
}

// @title System Conf API
// @version 0.1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name token
// @description This is a http server for engineering vehicle management system.

// @BasePath /api/system
func main() {
	args := &Args{}
	flag.IntVar(&args.Port, "port", 8081, "service port")
	flag.Parse()
	engine := gin.Default()
	apiRoot := engine.Group("/api")

	ctrl := api.NewController(apiRoot)
	ctrl.AutoBindSystem()

	apiRoot.GET("/docs/*any", handleDocs)
	err := engine.Run(fmt.Sprintf(":%d", args.Port))
	if err != nil {
		log.Panic(err)
	}

}
