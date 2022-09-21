package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mdlayher/wol"
)

var macAddrRegex, ipAddrRegex *regexp.Regexp

var broadcastIP string
var sharedKey string = "DefaultPassword"
var listenPort string = "80"

func main() {
	// load env vars
	if os.Getenv("SHARED_KEY") != "" {
		sharedKey = os.Getenv("SHARED_KEY")
	}
	if os.Getenv("LISTEN_PORT") != "" {
		listenPort = os.Getenv("LISTEN_PORT")
	}

	macAddrRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	ipAddrRegex = regexp.MustCompile(`^(\d+\.\d+\.\d+\.)\d+$`)

	broadcastIP = GetBroadcastIP()
	router := chi.NewRouter()

	// A good base middleware stack
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(5 * time.Second))

	loginPageTemplate, err := template.New("Login Page").Parse(loginPage)
	if err != nil {
		log.Fatalln("Failed to parse login template")
	}

	router.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("shared-key")
		// show success page if logged in or delete invalid cookie
		if err == nil {
			if cookie.Value == sharedKey {
				//show success page
				loginPageTemplate.Execute(w, struct{ Success bool }{true})
			} else {
				//delete cookie
				http.SetCookie(w, &http.Cookie{Name: "shared-key", MaxAge: -1})
			}
		} else {
			//show login page
			loginPageTemplate.Execute(w, nil)
		}
	})

	router.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		//validate shared key entry
		userSharedKey := r.FormValue("shared-key")
		if userSharedKey != sharedKey {
			//delete cookie
			http.SetCookie(w, &http.Cookie{Name: "shared-key", MaxAge: -1})
			loginPageTemplate.Execute(w, struct{ Success, Failure bool }{false, true})
			return
		}
		//login succeeded
		cookie := &http.Cookie{
			Name:     "shared-key",
			Value:    userSharedKey,
			Expires:  time.Now().Add(time.Hour * 1),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)
		loginPageTemplate.Execute(w, struct{ Success bool }{true})
	})

	router.Get("/wakeup/{macAddr}", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("shared-key")
		if err != nil || cookie.Value != sharedKey {
			//redirect on bad or no auth cookie
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		macAddrBytes, err := base64.StdEncoding.DecodeString(chi.URLParam(r, "macAddr"))
		macAddr := string(macAddrBytes)
		if err == nil && macAddrRegex.Match(macAddrBytes) {
			wakeOnLan(string(macAddr))
			w.Write([]byte("Sent Magic Packet to " + broadcastIP + " with " + macAddr))
			log.Println("Sent Magic Packet to " + broadcastIP + " with " + macAddr)
		} else {
			http.Error(w, "fail decode mac address (base64)", http.StatusInternalServerError)
			log.Println("ERROR: fail decode mac address")
		}
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Magic Packet Server."))
	})

	fmt.Println("Starting Server. broadcasting on: " + broadcastIP + ", sharedKey is: " + sharedKey)
	http.ListenAndServe(":3333", router)
}

// wakeOnLan sends a magic packet to the provided mac address
func wakeOnLan(macAddr string) {
	client, err := wol.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	hwmac, err := net.ParseMAC(macAddr)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Wake(broadcastIP+":9", hwmac)
	if err != nil {
		log.Fatal(err)
	}
}

// GetBroadcastIP gets preferred outbound ip address
func GetBroadcastIP() string {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ipnet := ipAddrRegex.FindStringSubmatch(localAddr.IP.String())
	return ipnet[1] + "0"
}

var loginPage string = `
<h1>Login</h1>
<form method="POST">
		<label>shared-key:</label><br />
		<input type="text" name="shared-key"><br />
		<input type="submit">
</form>
{{if .Success}}
<h2>Logged In.</h2>
{{end}}
{{if .Failure}}
<h2>Login Failed</h2>
{{end}}
`
