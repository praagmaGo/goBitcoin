package main

import (
	"fmt"
	"net"
        "net/http"
	"log"
	"time"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
    //remPartOfURL := r.URL.Path[len("/hello/"):]
    fmt.Fprintf(w, "Welcome to onBlock")
    fmt.Println("--",r.Method," main   by ",r.RemoteAddr)
}

func infosHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "onBlock informations")
    fmt.Println("--",r.Method," infos  by ",r.RemoteAddr)
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
//    fmt.Fprintf(w, "onBlock informations")
    fmt.Println("--",r.Method," infos  by ",r.RemoteAddr)
}

func main() {
  interfs,err:=net.Interfaces()
  if err != nil {
    fmt.Println("--erreur--")
    log.Fatal(err)
  }
  fmt.Println("Interfaces:")
  for _,interf:=range interfs{
    fmt.Println(" ", interf.Index, ": ", interf.Name,"\t",
      interf.HardwareAddr," ",interf.Flags)
  }
  fmt.Println("--------------")

  t:=time.Now()
  fmt.Printf("Il est %s\n", t.Local())

/*  ticker := time.NewTicker(time.Millisecond * 5000)
    go func() {
        for t := range ticker.C {
            fmt.Println("Tick at", t)
        }
    }()
*/
    http.HandleFunc("/", mainHandler)
    http.HandleFunc("/infos/", infosHandler)
    http.HandleFunc("/receiver.php", probeHandler)
    http.ListenAndServe("localhost:8080", nil)

//    defer ticker.Stop();
    defer fmt.Println("Ticker stopped")
}

