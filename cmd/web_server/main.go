package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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
	log.Printf("Trying get port from environment...\n")
	if port == "" {
		port = "80"
		log.Printf("Defaulting to port %s", port)
	}

	router := mux.NewRouter()

	router.Handle("/exectime/{duration}", http.HandlerFunc(execTimeHandler))

	router.Handle("/", http.HandlerFunc(textHandler))

	router.Handle("/ws/echo", http.HandlerFunc(wsHandler))

	router.Handle("/status/{code}", http.HandlerFunc(statusHandler))

	router.Handle("/content/json", http.HandlerFunc(jsonHandler))
	router.Handle("/content/html", http.HandlerFunc(htmlHandler))

	router.Handle("/redirect", http.RedirectHandler("https://www.example.com", http.StatusFound))

	svr := http.Server{
		Addr: ":" + port,
	}
	svr.Handler = router

	log.Fatal(svr.ListenAndServe())
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code, err := strconv.Atoi(vars["code"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, fmt.Sprint("invalid status code"))
		return
	}

	statusText := http.StatusText(code)
	if statusText == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, fmt.Sprint("invalid status code"))
		return
	}

	w.WriteHeader(code)
	_, _ = fmt.Fprint(w, fmt.Sprintf("%d %s", code, statusText))
}

func execTimeHandler(w http.ResponseWriter, r *http.Request) {
	const MaxDuration = 120 * time.Second

	vars := mux.Vars(r)
	duration, err := time.ParseDuration(vars["duration"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, fmt.Sprint("invalid duration"))
		return
	}

	if duration > MaxDuration {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, fmt.Sprintf("execution time must under %0.0f second", MaxDuration.Seconds()))
		return
	}

	time.Sleep(duration)
	_, _ = fmt.Fprint(w, fmt.Sprintf("got duration %s", duration.String()))
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

var upgrader = websocket.Upgrader{} // use default options

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
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
