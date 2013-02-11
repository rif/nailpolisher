package main

import (
	"app/controllers"
	"app/models"
	"fmt"
	"github.com/dchest/captcha"
	"log"
	"net/http"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ip := os.Getenv("OPENSHIFT_INTERNAL_IP") // empty is localhost
	router := models.Router
	// static
	router.Add("GET", "/static/", http.FileServer(http.Dir(models.BASE_DIR))).Name("static")
	// auth
	router.Add("GET", "/login/", controllers.Handler(controllers.LoginForm)).Name("login")
	router.Add("POST", "/login/", controllers.Handler(controllers.Login))
	router.Add("GET", "/fblogin", controllers.Handler(controllers.FbLogin)).Name("fblogin")
	router.Add("GET", "/gllogin", controllers.Handler(controllers.GlLogin)).Name("gllogin")

	router.Add("GET", "/logout/{csrf_token:[0-9a-z]+}", controllers.Handler(controllers.Logout)).Name("logout")

	router.Add("GET", "/register/", controllers.Handler(controllers.RegisterForm)).Name("register")
	router.Add("POST", "/register/", controllers.Handler(controllers.Register))
	router.Add("GET", "/captcha/", captcha.Server(captcha.StdWidth, captcha.StdHeight))

	router.Add("GET", "/profile/", controllers.Handler(controllers.ProfileForm)).Name("profile")
	router.Add("POST", "/profile/", controllers.Handler(controllers.Profile))

	router.Add("GET", "/reset/", controllers.Handler(controllers.ResetPasswordForm)).Name("reset")
	router.Add("POST", "/reset/", controllers.Handler(controllers.ResetPassword))

	router.Add("GET", "/change/", controllers.Handler(controllers.ChangePasswordForm)).Name("change")
	router.Add("POST", "/change/", controllers.Handler(controllers.ChangePassword))

	router.Add("GET", "/changetoken/{uuid:[0-9a-z]+}", controllers.Handler(controllers.ChangePasswordTokenForm)).Name("change_token")
	router.Add("POST", "/changetoken/{uuid:[0-9a-z]+}", controllers.Handler(controllers.ChangePasswordToken))

	//contact
	router.Add("GET", "/contact/", controllers.Handler(controllers.ContactForm)).Name("contact")
	router.Add("POST", "/contact/", controllers.Handler(controllers.Contact))

	// static
	router.Add("GET", "/page/{p:[a-z]+}", controllers.Handler(controllers.Static)).Name("page")

	// language
	router.Add("GET", "/language/{lang:[a-z]{2}}", controllers.Handler(controllers.SetLanguage)).Name("language")

	// index
	router.Add("GET", "/", controllers.Handler(controllers.Index)).Name("index")

	log.Print("The server is listening...")
	if err := http.ListenAndServe(fmt.Sprintf("%s:8080", ip), router); err != nil {
		log.Print("cmo server: ", err)
	}
}
