package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mcmcx.com/gpt-server/utils"
)

const LOG_HTTP_PREFIX = "(HTTP)"

type Server struct {
	//
	address    string
	port       int
	https_port int
	// Additionally, files containing a certificate and
	// matching private key for the server must be provided
	https_privatekey  string
	https_certificate string

	//
	router *gin.Engine
}

func InitServer(config Config, mode string) *Server {

	//
	gin.SetMode(mode)

	//
	router := gin.Default()
	server := Server{
		router: router,
		//
		address:           "",
		port:              -1,
		https_port:        -1,
		https_privatekey:  "",
		https_certificate: "",
	}

	server.address = config.Address
	if server.address == "" {
		server.address = "0.0.0.0"
	}
	if config.IPv6 {
		server.address = "[::]"
	}

	server.port = config.Port
	server.https_port = config.HTTPSPort
	server.https_certificate = config.HTTPSCertificate
	server.https_privatekey = config.HTTPSPrivateKey

	// custom logs
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		var text = fmt.Sprintf("[%s] (%s) | %s \"%s\" (%s, %d, %s)",
			param.TimeStamp.Format(time.RFC3339),
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			//param.Request.UserAgent(),
			//param.ErrorMessage,
		)

		utils.Logger.Log(LOG_HTTP_PREFIX, text)
		if len(param.ErrorMessage) > 0 {
			utils.Logger.LogError(LOG_HTTP_PREFIX, "Error Message : ", param.ErrorMessage)
		}
		return text + "\n"
	}))

	//
	api := router.Group("/api")
	api.Static("/assets", "assets")

	//
	register_handlers(router)

	//
	return &server
}

func (self Server) StartHTTPServer() bool {
	var port = -1
	if self.port > 0 {
		port = self.port
	}
	if port < 0 {
		utils.Logger.LogWarning(LOG_HTTP_PREFIX, "HTTP Server closed")
		return false
	}

	go start_http_server(self.router, fmt.Sprintf("%s:%d", self.address, port))

	//
	defer utils.Logger.LogWarning(LOG_HTTP_PREFIX, "HTTP Server starting on ", port)
	return true
}

func (self Server) StartHTTPSServer() bool {
	var port = -1
	if self.https_port > 0 {
		port = self.https_port
	}
	if port < 0 || self.https_privatekey == "" || self.https_certificate == "" {
		utils.Logger.LogWarning(LOG_HTTP_PREFIX, "HTTPS Server closed")
		return false
	}

	go start_https_server(self.router, fmt.Sprintf("%s:%d", self.address, port), self.https_certificate, self.https_privatekey)

	//
	defer utils.Logger.LogWarning(LOG_HTTP_PREFIX, "HTTPS Server starting on ", port)
	return true
}

// IPv6
// router.Run(":8080") // listen and serve on 0.0.0.0:8080
func start_http_server(router *gin.Engine, address string) bool {
	var err = http.ListenAndServe(address, router)
	if err != nil {
		utils.Logger.LogError(LOG_HTTP_PREFIX, "HTTP Error: ", err.Error(), "")
		return false
	}
	return true
}

func start_https_server(router *gin.Engine, address string, cert_file string, key_file string) bool {
	var err = http.ListenAndServeTLS(address,
		cert_file, key_file,
		router)
	if err != nil {
		utils.Logger.LogError(LOG_HTTP_PREFIX, "HTTPS Error: ", err.Error(), "")
		return false
	}
	return true
}

func register_handlers(router *gin.Engine) bool {
	//
	router.GET("/ping", HandlePing)

	// OpenAI API
	router.Any("/v1/models", HandleOpenAIModels)
	router.POST("/v1/chat/completions", HandleOpenAICompletions)

	//
	return true
}