package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type information struct {
	WelcomeMsg string
	Time       string
	HostIP     string
}

func newInformation() information {
	return information{
		WelcomeMsg: "Not Welcome - Develop Server",
		Time:       time.Now().Format("15:04:05"),
		HostIP:     GetOutboundIP().String(),
	}
}

func main() {
	// usage: PORT=8899 go run cmd/web_server/main.go
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
		log.Printf("Defaulting to port %s", port)
	}

	router := mux.NewRouter()

	router.Handle("/json/", http.HandlerFunc(jsonHandler))
	router.Handle("/html/", http.HandlerFunc(htmlHandler))
	router.Handle("/", http.HandlerFunc(textHandler))
	router.Handle("/redirect", http.RedirectHandler("http://www.example.com", http.StatusFound))

	svr := http.Server{
		Addr: ":" + port,
	}
	svr.Handler = router

	log.Fatal(svr.ListenAndServe())
}

func textHandler(w http.ResponseWriter, r *http.Request) {
	i := newInformation()

	_, _ = fmt.Fprint(w, fmt.Sprint(i))
}

func htmlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	i := newInformation()

	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal("template parse error, ", err)
	}
	_ = tmpl.Execute(w, i)
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	marshal, _ := json.Marshal(newInformation())

	w.Header().Add("Content-Type", "application/json")

	_, _ = fmt.Fprint(w, string(marshal))
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
