package main

import (
  "html/template"
//  "io/ioutil"
  "net/http"
//  "net/url"
  "fmt"
  "time"
  "strings"
//  "log"
//  "reflect"
  "github.com/gorilla/websocket"
  "reflect"
  "github.com/lib/pq"
  "database/sql"
  "os"
  "os/signal"
  "syscall"
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

type pageRT struct {
  rtChann chan string
  rtId int
}

var (
	db	*sql.DB
  err *pq.Error

	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
  rtChann chan string
  tabChan [] pageRT
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
      fmt.Println(err)    //on a probablement err qui affiche "EOF"

      wsEntryId:=tabChan[chanNum].rtId
      tabChan[chanNum].rtId=0
      fmt.Println("Fermeture ws wsEntryId: ",wsEntryId)
      db.QueryRow("insert into WebAccessLog (page,goin) values ('wsOut',$1)",wsEntryId)
			break
		}
	}
}

func writer(ws *websocket.Conn,chanNum int) {
//	lastError := ""
//	hello := "toto"
  var tunnel chan string

  tunnel=tabChan[chanNum].rtChann

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
      //  fmt.Printf("rt chan values [%d] %s\n",ret,Values)
        if (!ret) {	// on avait (!ret)
          fmt.Println("rt channel closed")
        } else {
      //    fmt.Println("rt receiving: ", Values)
          p:=[]byte(Values)

          if p != nil {
            ws.SetWriteDeadline(time.Now().Add(writeWait))
            if err := ws.WriteMessage(websocket.TextMessage,p); err != nil {
              fmt.Println("Error writing ws")
              fmt.Print(err)
              tabChan[chanNum].rtId=0
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
  fmt.Println("entering websocket");
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			//log.Println(err)
      fmt.Println(err)
		}
		return
	}

//	var lastMod time.Time
//	if n, err := strconv.ParseInt(r.FormValue("lastMod"), 16, 64); err != nil {
//		lastMod = time.Unix(0, n)
//	}

    fmt.Println("starting websocket");

    tunnel:=make(chan string)
    rt:=pageRT{rtChann:tunnel,rtId:1}

    var iTotal=len(tabChan)
    var chanNum int
    fmt.Println("Total tunnel entree",iTotal)
    var bAdd=true
    for idx,chemin := range tabChan {
      if chemin.rtId==0{
        fmt.Println("Chemin entre dans ",idx)
        tabChan[idx]=rt
        chanNum=idx
        bAdd=false
        break
      }
    }
    if bAdd{
      tabChan=append(tabChan,rt)
      chanNum=iTotal
    }
    iTotal=len(tabChan)
    fmt.Println("Total tunnel sortie",iTotal)

//-------------------------WebAccessLog
  moiIp:=r.RemoteAddr
  moiIp = moiIp[:strings.Index(moiIp,":")]
  var wsEntryId int
  db.QueryRow("insert into WebAccessLog (page,who4) values ('wsIn',$1) returning id",
    moiIp).Scan(&wsEntryId)
  fmt.Println("Ouverture ws wsEntryId: ",wsEntryId)
  tabChan[chanNum].rtId=wsEntryId

	go writer(ws,chanNum)
	reader(ws,chanNum)
}

type Ligne struct {
  Valeur1 string
  Valeur2 string
}

type rtLigne struct {
  Hhmm string
  Who4 string
  Duree string
}

type DataRoot struct {
	Title string
	Body  []byte
	Tableau []string
	Heure string
	Lignes []*Ligne
	Host	string
	Data	string
}

type DataWebLog struct {
	Title string
	Heure string
  CpteChan int
  NbEnCours int
	EnCours []*rtLigne
  NbHistos int
	Histos []*rtLigne
}
/*
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
*/
func rootHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("--",r.Method," root   by ",r.RemoteAddr)

	title:="onBlock"
//-------------------------WebAccessLog
  moiIp:=r.RemoteAddr
  moiIp = moiIp[:strings.Index(moiIp,":")]
  var pageEntryId int
  err:=db.QueryRow("insert into WebAccessLog (page,who4) values ('root',$1) returning id",
    moiIp).Scan(&pageEntryId)
  fmt.Println("pageEntryId: ",pageEntryId)

//-------------------------db access
  var numrows int
  err = db.QueryRow("select count(*) from Data_Continents").Scan(&numrows)
  if err != nil {
    fmt.Printf("sql.counting error: %v\n",err)
    return
  }
  fmt.Println("Nombre de colonnes:",numrows)

//get rows
  rows, err := db.Query("select * from Data_Continents")
  if err != nil {
    fmt.Println("typeOfErr:",reflect.TypeOf(err)) //  *pq.Error
    fmt.Printf("sql.select error: %v\n",err)
    return
  }
  defer rows.Close()

  t:=time.Now()

  p:=&DataRoot{Title:title,Heure:t.String()}
  p.Lignes=make([]*Ligne,numrows)
  fmt.Println("Nombre de lignes: ",len(p.Lignes))

  cpte:=0
  for rows.Next()  {    // rows est de type *sql.Rows
	  var alpha2 string
    var nom string

    err = rows.Scan(&alpha2, &nom)
    if err != nil {
      fmt.Printf("rows.Scan error: %v\n",err)
      return
    }
    l:=Ligne{Valeur1:alpha2,Valeur2:nom}
    p.Lignes[cpte]=&l
    cpte++

//    fmt.Printf("ct2: %v ctLong: %v \n",ct2, ctLong);
  }
/*
    for _,row := range p.Lignes{	//OK
      fmt.Println(reflect.TypeOf(row))
      fmt.Println(row.Valeur1,"  ",row.Valeur2)
    }
*/
	p.Tableau=[]string{"AAA","BBB","CCC"}

//        fmt.Println("Just before render")
//	renderTemplate(w, "view", p)
	p.Host=r.Host
	p.Data="Real time"
//	tpl, _ := template.ParseFiles("view.html")
	tpl, _ := template.New("foo").Parse(pageView)
	tpl.Execute(w, p)
  fmt.Println("Sortie root")
}

func weblogHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--",r.Method," weblog  by ",r.RemoteAddr)
//-------------------------WebAccessLog
  moiIp:=r.RemoteAddr
  moiIp = moiIp[:strings.Index(moiIp,":")]
  db.QueryRow("insert into WebAccessLog (page,who4) values ('weblog',$1) returning id",
    moiIp)
//-------------------------Données
  CpteChan:=0
  for _,chemin := range tabChan {
    if chemin.rtId!=0{
      CpteChan++
    }
  }
  t:=time.Now()
  p:=&DataWebLog{Title:"WebLog",Heure:t.String(),CpteChan:CpteChan}
  fmt.Println("Nb chann ouverts :",CpteChan)

//------------------------- tableau encours RT
  var numrows int
  request:="select count(d.id) from webaccesslog d left outer join webaccesslog f on d.id = f.goin where d.page='wsIn' and f.id is null"
  errtoto := db.QueryRow(request).Scan(&numrows)
//  fmt.Println("typeOfErr: errtoto [",reflect.TypeOf(errtoto),"]") //  *errors.errorString
  if errtoto != nil {
    fmt.Printf("sql.counting error: %v\n",err)
    return
  }
  fmt.Println("Total rt encours :",numrows)
  p.NbEnCours=numrows

  if numrows > 0 {
    request="select d.hhmm as debut, now()-d.hhmm as duree,d.who4 as origin from webaccesslog d left outer join webaccesslog f on d.id = f.goin where d.page='wsIn' and f.id is null order by debut desc"
    rows, err := db.Query(request)
    if err != nil {
      fmt.Println("typeOfErr:",reflect.TypeOf(err)) //  *pq.Error
      fmt.Printf("sql.select error: %v\n",err)
      return
    }
    defer rows.Close()
    p.EnCours=make([]*rtLigne,numrows)
    fmt.Println("Lectures EnCours: ",len(p.EnCours))

    cpte:=0
    for rows.Next()  {    // rows est de type *sql.Rows
	    var hhmm time.Time
      var duree string
      var who4 string

      err = rows.Scan(&hhmm,&duree,&who4)
      if err != nil {
        fmt.Printf("rows.Scan error: %v\n",err)
        return
      }

      l:=rtLigne{Hhmm:hhmm.String(),Who4:who4,Duree:duree}
      p.EnCours[cpte]=&l
      cpte++
    }
  }
//------------------------- tableau histos
  //var errtoto *sql.Row

  request="select count(d.hhmm) from webaccesslog d inner join webaccesslog f on d.id = f.goin"
  errtoto = db.QueryRow(request).Scan(&numrows)
//  fmt.Println("typeOfErr: errtoto [",reflect.TypeOf(errtoto),"]") //  *errors.errorString
  if errtoto != nil {
    fmt.Printf("sql.counting error: %v\n",err)
    return
  }
  fmt.Println("Total rt histos  :",numrows)
  p.NbHistos=numrows

  if numrows > 0 {
    request="select d.hhmm as debut,d.who4, f.hhmm-d.hhmm as duree from webaccesslog d inner join webaccesslog f on d.id = f.goin order by debut desc"
    rows, err := db.Query(request)
    if err != nil {
      fmt.Println("typeOfErr:",reflect.TypeOf(err)) //  *pq.Error
      fmt.Printf("sql.select error: %v\n",err)
      return
    }
    defer rows.Close()
    p.Histos=make([]*rtLigne,numrows)
    fmt.Println("Lectures Histos:",len(p.Histos))

    cpte:=0
    for rows.Next()  {    // rows est de type *sql.Rows
	    var hhmm time.Time
      var who4 string
      var duree string

      err = rows.Scan(&hhmm, &who4,&duree)
      if err != nil {
        fmt.Printf("rows.Scan error: %v\n",err)
        return
      }

      l:=rtLigne{Hhmm:hhmm.String(),Who4:who4,Duree:duree}
      p.Histos[cpte]=&l
      cpte++
    }
  }
//-------------------------Affichage
//	tpl, _ := template.ParseFiles("weblog.html")
	tpl, _ := template.New("foo").Parse(pageWebLog)
	tpl.Execute(w, p)
  fmt.Println("Sortie weblog")
}                                                                                                                                                                                                                                            

func jaugeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--",r.Method," weblog  by ",r.RemoteAddr)
//-------------------------WebAccessLog
  moiIp:=r.RemoteAddr
  moiIp = moiIp[:strings.Index(moiIp,":")]
  db.QueryRow("insert into WebAccessLog (page,who4) values ('jauge',$1) returning id",
    moiIp)
//	tpl, _ := template.ParseFiles("jauge.html")
	tpl, _ := template.New("foo").Parse(pageJauge)
	tpl.Execute(w, nil)
  fmt.Println("Sortie root")
}

/*
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
		p = &DataRoot{Title: title,Heure:t.String()}
		p.Body=[]byte("troutisme")
	}
	//renderTemplate(w, "edit", p)
	tpl, _ := template.ParseFiles("edit.html")
	tpl.Execute(w, p)
}
*/
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

  for _,chemin := range tabChan {
    if chemin.rtId!=0{
      chemin.rtChann<-ligne
//      tabChan[idx]<-ligne
    }
  }
  fmt.Println("probe_in exit")

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
  db2, err2 := sql.Open("postgres", "dbname=goumzou user=ubuntu password=toto sslmode=disable")
    if err2 != nil {
      fmt.Printf("sql.Open error: %v\n",err)
      return
    }
  defer db.Close()
  db=db2
  fmt.Println(" - OK")

//-------------------------testing allocation
//    var test []int
//    test=make([]int, 4)
//    fmt.Println("Taille de test: ",len(test))

//-------------------------signal

  signalChan := make(chan os.Signal, 1)
  signal.Notify(signalChan, os.Interrupt,syscall.SIGTERM)
  go func() {
 //   var retour *sql.Row // source code: return &Row{rows: rows, err: err}

//    <-signalChan
    for sig := range signalChan {
      fmt.Printf("Got signal \"%v\", cleaning and stopping...\n", sig)
//cleaning
      CpteChan:=0
      for _,chemin := range tabChan {
        if chemin.rtId!=0{
          CpteChan++
          wsEntryId:=chemin.rtId
          chemin.rtId=0
          db.QueryRow("insert into WebAccessLog (page,goin) values ('term',$1)",wsEntryId)
    //      switch{
    //        case retour != nil:
    //          fmt.Printf("Insert fatal\n")
    //        default:
    //          fmt.Printf("Insert OK\n")
    //      }
          close(chemin.rtChann)
        }
      }
      fmt.Println("Nb fermetures: ",CpteChan)
      os.Exit(1)
    }
  }()

//-------------------------init channel
    rtChann=make(chan string)

//-------------------------init ws table
//  var tabChan []chan string
  tabChan=make([]pageRT,0)

//-------------------------html server
  fmt.Print("starting web server")
	http.HandleFunc("/ws", serveWs)
  http.HandleFunc("/probe_in", probeHandler)
  http.HandleFunc("/", rootHandler)
  http.HandleFunc("/weblog", weblogHandler)
  http.HandleFunc("/jauge", jaugeHandler)
//  http.HandleFunc("/save/", saveHandler)
//  http.HandleFunc("/test/", testHandler)
  fmt.Println(" - OK")
  http.ListenAndServe(":8080", nil)
}

