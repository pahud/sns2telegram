package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/apex/gateway"
	"github.com/gin-gonic/gin"
)

var tg *Telegram

var baseURL string

// request body byte
var reqbodybyte []byte

// Chat ...
type Chat struct {
	ID    int32  `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// Msg ...
type Msg struct {
	MessageID int32                  `json:"message_id"`
	From      map[string]interface{} `json:"from"`
	Chat      Chat                   `json:"chat"`
	Date      int32                  `json:"date"`
	Text      string                 `json:"text"`
}

// Req ...
type Req struct {
	Message  Msg   `json:"message"`
	UpdateID int64 `json:"update_id"`
}

func init() {
	telegramToken := os.Getenv("telegram_token")

	tg = NewTelegram()
	// change this with your Telegram bot token
	// tg.token = "123456789:xxxxxxxxxxxxxxxxxxx"
	tg.token = telegramToken
}

func int32tostr(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}

func inLambda() bool {
	if lambdaTaskRoot := os.Getenv("LAMBDA_TASK_ROOT"); lambdaTaskRoot != "" {
		return true
	}
	return false
}

func helloHandler(c *gin.Context) {
	name := c.Param("name")
	c.String(http.StatusOK, "Hello %s", name)
}

func welcomeHandler(c *gin.Context) {
	name := c.Param("name")
	city := c.Param("city")
	c.String(http.StatusOK, "Welcome %v from %v\n", name, city)
}

func telegramCbHandler(c *gin.Context) {
	var err error
	reqbodybyte, _ = c.GetRawData()
	var req Req
	err = json.Unmarshal(reqbodybyte, &req)
	if err != nil {
		log.Printf("json.Unmarshal got error:%v", err)
		c.JSON(http.StatusOK, gin.H{
			"message": "got error",
		})
		return
	}
	text := req.Message.Text
	chatID := req.Message.Chat.ID
	log.Printf("got text:%v", text)
	if strings.HasPrefix(text, "/init") {
		// initialization
		log.Printf("got /init, start initHandler()")
		initHandler(c)
		return
	}
	tg.sendMessage("unknown command", int32tostr(chatID))
	c.JSON(http.StatusOK, gin.H{
		"message": "unknown command",
	})

}

func webhookHandler(c *gin.Context) {
	ignoreAttributes := [4]string{
		"Signature",
		"SignatureVersion",
		"SigningCertURL",
		"UnsubscribeURL",
	}
	strchatID := c.Param("chatID")
	UUID := c.Param("UUID")
	db, err := NewDdbHandler()
	if err != nil {
		log.Printf("got error:%v", err)
		c.JSON(http.StatusOK, gin.H{
			"message": "got error",
		})
		return
	}
	i, err := strconv.ParseInt(strchatID, 10, 32)
	if err != nil {
		panic(err)
	}
	chatID := int32(i)
	if db.IsValidChatID(chatID, UUID) {
		log.Println("valid chatID")
		reqbodybyte, _ = c.GetRawData()
		var reqbodyobj map[string]interface{}
		err = json.Unmarshal(reqbodybyte, &reqbodyobj)
		// if Type=='SubscriptionConfirmation'
		if reqbodyobj["Type"] == "SubscriptionConfirmation" {
			subURL := reqbodyobj["SubscribeURL"].(string)
			msg2send := "click the following subscribe URL to confirm: \n\n" + subURL
			tg.sendMessage(msg2send, int32tostr(chatID))
			c.JSON(http.StatusOK, gin.H{
				"message": "message sent",
			})
			return
		}
		if strings.HasPrefix(reqbodyobj["Message"].(string), `{`) {
			// Message contains json
			log.Info("Message contains json payload, decoding it")
			var msgobj map[string]interface{}
			err = json.Unmarshal([]byte(reqbodyobj["Message"].(string)), &msgobj)
			if err != nil {
				log.Errorf("json.Unmarshal got err: %v", err)
				return
			}
			reqbodyobj["Message"] = msgobj
		}
		// remove ignoreAttributes
		for i := range ignoreAttributes {
			delete(reqbodyobj, ignoreAttributes[i])
		}

		resp, _ := json.MarshalIndent(reqbodyobj, "", "   ")
		tg.sendMessage(string(resp), strchatID)
		c.JSON(http.StatusOK, gin.H{
			"message": "message sent",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Invalid ChatID or UUID",
		})
	}
}

func initHandler(c *gin.Context) {
	var err error
	db, err := NewDdbHandler()
	if err != nil {
		log.Printf("got error:%v", err)
		c.JSON(http.StatusOK, gin.H{
			"message": "got error",
		})
		return
	}

	var req Req
	err = json.Unmarshal(reqbodybyte, &req)
	if err != nil {
		log.Printf("json.Unmarshal got error:%v", err)
		c.JSON(http.StatusOK, gin.H{
			"message": "got error",
		})
		return
	}
	log.Infof("ChatID: %v", req.Message.Chat.ID)
	chatID := req.Message.Chat.ID
	var strChatID string
	var UUID string

	strChatID, UUID, err = db.CreateItem(chatID)
	if err != nil {
		log.Printf("got error:%v", err)
		c.JSON(http.StatusOK, gin.H{
			"message": "got error",
		})
		return
	}

	baseURL := os.Getenv("base_url")
	webhookURL := baseURL + "/webhook/" + strChatID + "/" + UUID
	txt2send := "Your webhook endpoint:\n\n" + webhookURL + "\n\n*subscribe your SNS topic with this URL to receive SNS notifications in this chatroom.\n"
	tg.sendMessage(txt2send, strChatID)
	c.String(http.StatusOK, txt2send)
}

func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"text": "Welcome to gin lambda server.",
	})
}

func routerEngine() *gin.Engine {
	// set server mode
	gin.SetMode(gin.DebugMode)

	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health/ping", pingHandler)
	r.GET("/healthz", pingHandler)
	r.POST("/init", initHandler)
	r.POST("/webhook/:chatID/:UUID", webhookHandler)
	r.POST("/telegram/cb", telegramCbHandler)
	r.GET("/welcome/:name/:city", welcomeHandler)
	r.GET("/user/:name", helloHandler)
	r.GET("/", rootHandler)

	return r
}

func main() {
	baseURL := os.Getenv("base_url")

	if baseURL == "" {
		log.Fatal("base_url not found in the environment variables")
	}
	var addr string
	if addr = ":" + os.Getenv("PORT"); addr != ":" {
		log.Printf("listening on %v", addr)
	} else {
		addr = ":8888"
		log.Printf("listening on %v", addr)
	}
	if inLambda() {
		log.Fatal(gateway.ListenAndServe(addr, routerEngine()))
	} else {
		log.Fatal(http.ListenAndServe(addr, routerEngine()))
	}
}
