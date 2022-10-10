package main

import (
  "html/template"
  "io/ioutil"
  "net/http"
  "net/url"
  "fmt"
  "time"
  "strings"
  "github.com/ziutek/mymysql/mysql"
   _ "github.com/ziutek/mymysql/native" // Native engine
//  _ "github.com/ziutek/mymysql/thrsafe" // Thread safe engine
  "log"
//  "reflect"
  "github.com/gorilla/websocket"
  "reflect"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Poll file for changes with this period.
	filePeriod = 1 * time.Second
)

var (
	db	mysql.Conn
//	addr      = flag.String("addr", ":8080", "http service address")
//	homeTempl = template.Must(template.New("").Parse(homeHTML))
//	filename  string
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
  rtChann chan string
  tabChan []chan string
)

func reader(ws *websocket.Conn,chanNum int) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
      fmt.Println("Error reading ws")
      fmt.Println(err)
      //on a probablement err qui affiche "EOF"
      tabChan[chanNum]=nil
			break
		}
	}
}

func writer(ws *websocket.Conn,chanNum int) {
//	lastError := ""
//	hello := "toto"
  var tunnel chan string

  tunnel=tabChan[chanNum]

	pingTicker := time.NewTicker(pingPeriod)
//	fileTicker := time.NewTicker(filePeriod)
	defer func() {
		pingTicker.Stop()
//		fileTicker.Stop()
		ws.Close()
	}()
	for {
		select {
      case Values,ret:= <-tunnel:
        fmt.Printf("rt chan values [%d] %s\n",ret,Values)
        if (!ret) {	// on avait (!ret)
          fmt.Println("rt channel closed")
        } else {
          fmt.Println("rt receiving: ", Values)
          p:=[]byte(Values)

          if p != nil {
            ws.SetWriteDeadline(time.Now().Add(writeWait))
            if err := ws.WriteMessage(websocket.TextMessage,p); err != nil {
              fmt.Println("Error writing ws")
              fmt.Print(err)
              tabChan[chanNum]=nil
		          return
		        }
          }
        }
	/*	case <-fileTicker.C:
			var p []byte
			var err error

			fmt.Println("DataSending\n")
			bDataReady=false

		//	p, lastMod, err = readFileIfModified(lastMod)
			p=[]byte(hello)
			if hello=="toto" {hello="mumu"} else {hello="toto"}

			if err != nil {
				if s := err.Error(); s != lastError {
					lastError = s
					p = []byte(lastError)
				}
			} else {
				lastError = ""
			}

			if p != nil {
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteMessage(websocket.TextMessage, p); err != nil {
					return
				}
			}*/
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

//	var lastMod time.Time
//	if n, err := strconv.ParseInt(r.FormValue("lastMod"), 16, 64); err != nil {
//		lastMod = time.Unix(0, n)
//	}

    fmt.Println("starting websocket");

    tunnel:=make(chan string)
    var iTotal=len(tabChan)
    var chanNum int
    fmt.Println("Total tunnel entree",iTotal)
    var bAdd=true
    for idx,chemin := range tabChan {
      if chemin==nil{
        fmt.Println("Chemin entre dans ",idx)
        tabChan[idx]=tunnel
        chanNum=idx
        bAdd=false
        break
      }
    }
    if bAdd{
    //tabChan[iTotal]=make(chan string)
    tabChan=append(tabChan,tunnel)
    chanNum=iTotal
    }
    iTotal=len(tabChan)
    fmt.Println("Total tunnel sortie",iTotal)

	go writer(ws,chanNum)
	reader(ws,chanNum)
}

type Ligne struct {
  Valeur1 string
  Valeur2 string
}

type Page struct {
	Title string
	Body  []byte
	Tableau []string
	Heure string
	Lignes []*Ligne
	Host	string
	Data	string
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, p)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("--",r.Method," root   by ",r.RemoteAddr)
  fmt.Println("RequestURI: ",r.RequestURI);	//bon
//	title := r.URL.Path[len("/view/"):]
	title:="onBlock"
//-------------------------db access
    rows, res, err := db.Query("select * from Data_Continents")
    if err != nil {
        panic(err)
    }
    numrows:=len(rows)
    fmt.Println("Nombre de lignes: ",numrows)
    premier := res.Map("Continent_Alpha2")
    second := res.Map("Continent_Name")

    t:=time.Now()

    p:=&Page{Title:title,Heure:t.String()}
    p.Lignes=make([]*Ligne,numrows)
    fmt.Println("Nombre de lignes: ",len(p.Lignes))

    cpte:=0    
    for _, row := range rows {
//      alpha2 := row.Str(0)
      alpha2:=row.Str(premier)
      nom:=row.Str(second)
//      fmt.Println(alpha2,"  ",nom)
      l:=Ligne{Valeur1:alpha2,Valeur2:nom}
      p.Lignes[cpte]=&l
      cpte++
    }

//    for _,row := range p.Lignes{	//OK
//      fmt.Println(reflect.TypeOf(row))
//      fmt.Println(row.Valeur1,"  ",row.Valeur2)
//    }

//	p, err := loadPage(title)
//	if err==nil{
//		fmt.Println("view error loading")
//	}

	p.Tableau=[]string{"AAA","BBB","CCC"}

//        fmt.Println("Just before render")
//	renderTemplate(w, "view", p)
	p.Host=r.Host
	p.Data="Real time"
	tpl, _ := template.ParseFiles("view.html")
	tpl.Execute(w, p)

}

func editHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--",r.Method," edit   by ",r.RemoteAddr)
  fmt.Println("UrlString: ",r.URL.String())	//bon
  fmt.Println("RequestURI: ",r.RequestURI);	//bon
  fmt.Println("RawQuery: ",r.URL.RawQuery);	// slt les params en string

  u, err := url.Parse( r.URL.String())
  if err != nil {
    panic(err)
  }
  m := u.Query()
  fmt.Println(m)
  fmt.Println(len(m))
  fmt.Println(reflect.TypeOf(m)) // url.Values, qui est: map[string][]string
//  valeur1:=http.FormValue("toto")
//  fmt.Println(valeur1)

// deux valeurs suivantes sont bonnes
//  valeur1:=m["toto"][0]	// type string
//  fmt.Println("toto=",valeur1)

//  fmt.Println(reflect.TypeOf(valeur1))
  fmt.Println("finDecode")

  //param1:=m[“PeerIp”][0]

//	title := r.URL.Path[len("/edit/"):]
        title:="edition"
	p, err := loadPage(title)
	if err != nil {
		t:=time.Now()
		p = &Page{Title: title,Heure:t.String()}
		p.Body=[]byte("troutisme")
	}
	//renderTemplate(w, "edit", p)
	tpl, _ := template.ParseFiles("edit.html")
	tpl.Execute(w, p)
}

func testHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("--",r.Method," test       by ",r.RemoteAddr)
//  rtChann<-"kupuku"	// attention ca bloque si la ws n'est pas ouverte
//  fmt.Println("got: ",r.URL.RequestURI());
//  fmt.Println("got: ",r.URL.RawQuery);
//  param:=r.URL.Query()[“PeerIp”][0];
//  fmt.Println("param1: ",param);
  fmt.Println("UrlString: ",r.URL.String())	//bon
  fmt.Println("RequestURI: ",r.RequestURI);	//bon
  fmt.Println("RawQuery: ",r.URL.RawQuery);	// slt les params en string
  body := r.FormValue("body")
  fmt.Println("bodyValeur: ",body)
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("--",r.Method," probe_in   by ",r.RemoteAddr)
//  rtChann<-"kupuku"	// attention ca bloque si la ws n'est pas ouverte

//  fmt.Println("UrlString: ",r.URL.String())	//bon
//  fmt.Println("RequestURI: ",r.RequestURI);	//bon
//  fmt.Println("RawQuery: ",r.URL.RawQuery);	// slt les params en string
  pram1 := r.FormValue("Origin")
  pram1 = pram1[:strings.Index(pram1,":")]
  pram2 := r.FormValue("TotalVout")
  pram3 := r.FormValue("Time")
  pram3 = pram3[11:22]
  fmt.Println("Update: ",pram1," ",pram2," #",pram3)
//  fmt.Println("TotalVout: ",pram2)

//  pram2+="</td><td>tutu"
  ligne:=pram3+"</td><td>"+pram2+"</td><td>"+pram1

  for idx,chemin := range tabChan {
    if chemin!=nil{
      tabChan[idx]<-ligne
    }
  }


/*
  u, err := url.Parse( r.RequestURI)
  if err != nil {
    panic(err)
  }
  m := u.Query()
  fmt.Println(m)

//  m, _ := url.ParseQuery(u.RawQuery)
  fmt.Println(m)
  fmt.Println(len(m))
  fmt.Println("avec m: ",m["toto"])
//  param:=m[“PeerIp”]
  fmt.Println(reflect.TypeOf(m))
//  fmt.Println("param1: ",param);
  fmt.Println("finDecode")	*/
}

func main() {
//-------------------------db init
  fmt.Print("starting db connection")
    db = mysql.New("tcp", "", "127.0.0.1:3306","serge", "toto", "gromok")
    err := db.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    err = db.Ping()
    if err != nil {
      panic(err.Error()) // proper error handling instead of panic in your app
    }
    fmt.Println(" - OK")

//-------------------------testing allocation
//    var test []int
//    test=make([]int, 4)
//    fmt.Println("Taille de test: ",len(test))

//-------------------------init channel
    rtChann=make(chan string)

//-------------------------init ws table
//  var tabChan []chan string
  tabChan=make([]chan string,0)

//-------------------------html server
  fmt.Print("starting web server")
  http.HandleFunc("/probe_in", probeHandler)
  http.HandleFunc("/", rootHandler)
  http.HandleFunc("/edit/", editHandler)
//http.HandleFunc("/save/", saveHandler)
  http.HandleFunc("/test/", testHandler)
	http.HandleFunc("/ws", serveWs)
  fmt.Println(" - OK")
  http.ListenAndServe(":8080", nil)
}

