package main

import (
	"net/http"
	"net/url"
	"os"

	"math/rand"
	"time"

	"github.com/gin-gonic/gin"

	//1
	"github.com/gin-contrib/sessions"
	// "github.com/gin-contrib/sessions/cookie"
	//"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/redis"
)

type PageView struct {
	CurrentUser string
	PageTitle   string
	Title       string
	Text        string
}

type LoginView struct {
	CurrentUser string
	PageTitle   string
	Error       bool
	Email       string
}

func (v *LoginView) Validate() bool {
	if len(v.Email) < 3 {
		return false
	}
	return true
}

var theRandom *rand.Rand
var userkey = "SESSION_KEY_USERID"

func start(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	var currentUser = ""
	if user != nil {
		currentUser = user.(string)
	}
	computerName, _ := os.Hostname()
	c.HTML(http.StatusOK, "home.html",
		&PageView{CurrentUser: currentUser,
			PageTitle: "test",
			Title:     "Hej Golang",
			Text:      computerName})
}

func secretfunc(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	var currentUser = ""
	if user != nil {
		currentUser = user.(string)
	}
	c.HTML(http.StatusOK, "secret.html", &PageView{CurrentUser: currentUser, PageTitle: "test", Title: "Hej Golang", Text: "hejsan"})
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(userkey)
	session.Save()
	c.Redirect(302, "/")

}

func healthCheck(c *gin.Context) {
	// if redis == DEAD{
	// 	c.JSON(http.StatusInternalServerError,gin.H{})
	// }
	c.JSON(http.StatusOK, gin.H{})
}

func login(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", &LoginView{PageTitle: "Login"})
}
func loginPost(c *gin.Context) {
	var viewModel LoginView
	c.ShouldBind(&viewModel)
	if viewModel.Validate() {
		session := sessions.Default(c)
		session.Set(userkey, viewModel.Email)
		session.Save()
		redirectUrl := c.DefaultQuery("redirect_uri", "/")
		c.Redirect(302, redirectUrl)
		return
	}
	c.Status(200)
	c.HTML(http.StatusOK, "login.html", &viewModel)
}

var config Config

func main() {
	theRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
	readConfig(&config)

	// data.InitDatabase(config.Database.File,
	// 	config.Database.Server,
	// 	config.Database.Database,
	// 	config.Database.Username,
	// 	config.Database.Password,
	// 	config.Database.Port)

	router := gin.Default()

	//2
	var secret = []byte("secret")

	store, _ := redis.NewStore(10, "tcp", config.Redis.Server, "", secret)
	//store := memstore.NewStore([]byte(secret))
	router.Use(sessions.Sessions("mysession2", store))

	router.LoadHTMLGlob("templates/**")
	router.GET("/", start)
	//router.GET("/healthz", healthCheck)
	router.GET("/login", login)
	router.POST("/login", loginPost)
	router.GET("/logout", logout)

	//3 frivillig
	adminRoutes := router.Group("/admin")
	adminRoutes.Use(AuthRequired)
	adminRoutes.GET("/account", secretfunc)

	// loop initierar en databas tabell och skapar 30000 rader...
	// den tar 30 sekunder

	router.Run(":8080")
}

// 4 frivillig
func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	var redirectUrl = url.QueryEscape("http://" + c.Request.Host + c.Request.RequestURI)
	if user == nil {
		c.Redirect(302, "/login?redirect_uri="+redirectUrl)
		// Abort the request with the appropriate error code
		return
	}
	// Continue down the chain to handler etc
	c.Next()
}
