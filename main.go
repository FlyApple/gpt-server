package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"mcmcx.com/gpt-server/httpx"
	"mcmcx.com/gpt-server/server"
	"mcmcx.com/gpt-server/utils"
)

func main() {
	logger := utils.NewLogger()
	logger.Init()

	//
	bytes, err := os.ReadFile("config.yaml")
	if err != nil {
		logger.LogError("read file config.yaml error: ", err)
		return
	}

	var config server.Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		logger.LogError("parse config.yaml error: ", err)
		return
	}

	//
	server.API_IPInit()
	address, err := net.InterfaceAddrs()
	if address == nil {
		logger.LogError("Net interface error: ", err)
		return
	}
	ip := (address[0].(*net.IPNet)).IP
	ipi := server.IPLocalized(ip.String())
	logger.Log("IP ", ip.String(), " '", ipi.FullLocalize(), "'")

	//Redis
	if !server.RedisInitialize("config.yaml") {
		return
	}

	//ChatGPT
	if !server.API_GPTInit(config) {
		return
	}

	//Get ai models
	var data = server.API_GPTModels2()
	if data.ErrorCode != httpx.HTTP_RESULT_OK || !server.OpenAI_Init(data.Data()) {
		logger.LogError("GPT Loading models failure.")
		return
	}

	//
	logger.Log("GPT service loading ...")
	var service *server.Server = server.InitServer(config, gin.DebugMode)
	if service == nil {
		logger.LogError("[Server] Error: ", "init service error.")
		return
	}

	logger.Log("GPT service starting ...")
	service.StartHTTPServer()
	service.StartHTTPSServer()

	//select {}
	sigs := make(chan os.Signal, 1)
	//signal.Ignore(os.Interrupt)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	println("Signal -> %+v", sig)

	//
	println("Exiting ...")

	//
	server.RedisRelease()
	server.LDBReleaseAll()
	return
}
