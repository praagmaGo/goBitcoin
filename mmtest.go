package main

import (
  "fmt"
  "net/http"
  "html/template"
//  "database/sql"
//  _ "github.com/go-sql-driver/mysql"
  "github.com/ziutek/mymysql/mysql"
   _ "github.com/ziutek/mymysql/native" // Native engine
//  _ "github.com/ziutek/mymysql/thrsafe" // Thread safe engine
  "log"
  "time"
  "strconv"
  )
/*
var (
	db	mysql.Conn
)
*/

func icoHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("--",r.Method," icon   by ",r.RemoteAddr)
  fmt.Println("UrlString: ",r.URL.String())
  fmt.Println("Time: ",time.Now().Local());
  http.ServeFile(w,r,"favicon.ico")
  fmt.Println("Sortie icon")
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("--",r.Method," root   by ",r.RemoteAddr)
//  fmt.Println("UrlString: ",r.URL.String())	//bon
  fmt.Println("RequestURI: ",r.RequestURI);	//bon
  fmt.Println("Time: ",time.Now().Local());
  
  if r.RequestURI!="/"{
    fmt.Println("Sortie root 404")
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte("404 error"))
    return}

//	tpl, _ := template.ParseFiles("color.html")
	tpl, _ := template.New("foo").Parse(pageFastPay)
	tpl.Execute(w, nil)
//  fmt.Fprint(w, "You are inside the page")
  fmt.Println("Sortie root")
}

func evalHandler(w http.ResponseWriter, r *http.Request) {
  var	db	mysql.Conn

  fmt.Println("--",r.Method," eval   by ",r.RemoteAddr)
  fmt.Println("UrlString: ",r.URL.String())	//bon
//  fmt.Println("RequestURI: ",r.RequestURI);	//bon
//  fmt.Println("RawQuery: ",r.URL.RawQuery);	// slt les params en string
  fmt.Println("Time: ",time.Now().Local());

    db = mysql.New("tcp", "", "127.0.0.1:3306","serge", "toto", "gromok")
//    db = mysql.New("tcp", "", "127.0.0.1:3306","_SergeH", "UUjjnn56", "BitNodes2")
    err := db.Connect()
    if err != nil {
      fmt.Println("Erreur ouverture db")
      log.Fatal(err)
    }
    defer db.Close()
    err = db.Ping()
    if err != nil {
      panic(err.Error()) // proper error handling instead of panic in your app
    }

//  fmt.Fprint(w,"Evaluation... please wait")


//  duration:=time.Duration(4)*time.Second
//  time.Sleep(duration)

  var mess string

  pram1 := r.FormValue("HashId")
  long:=len(pram1)
  fmt.Println("pram1=",long," ",pram1)

  rows, res, err := db.Query("select DateTime,PeerId from Inputs_Peers where HashId='%s' order by DateTime asc",pram1)
//  fmt.Println("res=",res)
  if err != nil {
    fmt.Println("panic 01")
    mess="<fieldset>Query panic 1</fieldset>"
  } else if len(rows) == 0 {
    fmt.Println("no results")
    mess="<fieldset>Never seen in BTC P2P</fieldset>"
  } else{

    var Country string;
    var Region string;
    var City string;
    var DateTime string;
    var PeerId int;

    numrows:=len(rows)
    fmt.Println("Nombre de lignes: ",numrows) // "Reported by NB probes"
    premier := res.Map("DateTime")
    second := res.Map("PeerId")
    for _, row := range rows {
//      alpha2 := row.Str(0)
      fmt.Println("inside row")
      DateTime=row.Str(premier)
      PeerId=row.Int(second)
      fmt.Println("DateTime: ",DateTime)
      fmt.Println("PeerId  : ",PeerId)
      break;
    }
//  mess="<br><fieldset>mmtralala</fieldset>"
//   mess="<fieldset>Entered P2P on "+DateTime+"<br></fieldset>"

//-------------------------country
    rows, res, err := db.Query("select Country_Name from Data_Countries,Data_Peers where Data_Countries.Country_Alpha2 = Data_Peers.Country_Alpha2 and Data_Peers.PeerID=%d",PeerId)
    if res == nil {
      fmt.Println("no results")
      Country="CountryNotFound"
    } else{
      if err != nil {
  //        panic(err)
       fmt.Println("panic 02")
      }
      premier = res.Map("Country_Name")
      for _, row := range rows {
        Country=row.Str(premier)
        fmt.Println("Country:  ",Country)
        break;
      }
    }
//-------------------------region
    rows, res, err = db.Query("select Region_Name from Data_Regions,Data_Peers where Data_Regions.RegionID = Data_Peers.RegionID and Data_Peers.PeerID=%d",PeerId)
    if res == nil {
      fmt.Println("no results")
      Region="RegionNotFound"
    } else{
      if err != nil {
  //        panic(err)
       fmt.Println("panic 02")
      }
      premier = res.Map("Region_Name")
      for _, row := range rows {
//      alpha2 := row.Str(0)
        Region=row.Str(premier)
        fmt.Println("Region :  ",Region)
        break;
      }
    }
//-------------------------city
    rows, res, err = db.Query("select City_Name from Data_Cities,Data_Peers where Data_Cities.CityID = Data_Peers.CityID and Data_Peers.PeerID=%d",PeerId)
    if res == nil {
      fmt.Println("no results")
      City="CityNotFound"
    } else{
      if err != nil {
  //        panic(err)
       fmt.Println("panic 02")
      }
      premier = res.Map("City_Name")
      for _, row := range rows {
//      alpha2 := row.Str(0)
        City=row.Str(premier)
        fmt.Println("City   :  ",City)
        break;
      }
    }
    mess="<fieldset>Entered P2P on "+DateTime+"<br>  at "+City+","+Region+","+Country+"</fieldset>"
  }
// debut ajout
  mess2:="<br><fieldset>Check for money available</fieldset>"
// double spend
  mess3:="";
  var InHashID int;
  rows, res, err = db.Query("select InHashID from Good_InHashs where InHash='%s'",pram1)
  if err != nil {
    mess3="Query panic 3"
    fmt.Println(mess3)
  } else if len(rows) == 0 {
    fmt.Println("no results on Good_InHashs")
    mess3="No double spend detected on Good_InHashs"
  } else{
    numrows:=len(rows)
    fmt.Println("Nombre de lignes: ",numrows)
    premier := res.Map("InHashID")
    for _, row := range rows {
      InHashID=row.Int(premier)
      fmt.Println("InHashID: ",InHashID)
      break;
    }
    mess3="Double spend detected on Good_InHashs - InHashID="+strconv.Itoa(InHashID)
    fmt.Println(mess3)
  }
  mess3="<br><fieldset>"+mess3+"</fieldset>"

//  mess4:="<br><fieldset>Check how many peers the transaction is</fieldset>"
  mess4:=""
  var Number int;
  rows, res, err = db.Query("select count(*) as nbr from INV_TXT where HashINV=0x%s",pram1)
  if err != nil {
    mess4="Query panic 4"
    fmt.Println(mess4)
  } else if len(rows) == 0 {
    fmt.Println("no results on INV_TXT")
    mess3="Transaction not seen in any pool peers"
  } else{
    numrows:=len(rows)
    fmt.Println("Nombre de lignes: ",numrows)
    premier := res.Map("nbr")
    for _, row := range rows {
      Number=row.Int(premier)
      fmt.Println("Number: ",Number)
      break
    }
    mess4="Number of peers where the transaction is in: "+strconv.Itoa(Number)
    fmt.Println(mess4)
  }
  mess4="<br><fieldset>"+mess4+"</fieldset>"
// fin ajout

  fmt.Fprint(w,mess+mess2+mess3+mess4)
  fmt.Println("Sortie eval")
}

func main() {
  var	db	mysql.Conn

    fmt.Printf("------------------\n")

    db = mysql.New("tcp", "", "127.0.0.1:3306","serge", "toto", "gromok")
//    db = mysql.New("tcp", "", "127.0.0.1:3306","_SergeH", "UUjjnn56", "BitNodes2")
    err := db.Connect()
    if err != nil {
      fmt.Println("Erreur ouverture db")
      log.Fatal(err)
    }
    defer db.Close()
    err = db.Ping()
    if err != nil {
      panic(err.Error()) // proper error handling instead of panic in your app
    }/*
    stmtOut, err := db.Prepare("SELECT squareNumber FROM squarenum WHERE number = ?")
    if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
    }
    defer stmtOut.Close() */

// doc select dans https://github.com/ziutek/mymysql
//    rows, res, err := db.Query("select * from Data_Continents where id > %d", 20)
    rows, res, err := db.Query("select * from Data_Continents")
    if err != nil {
        panic(err)
    }
    premier := res.Map("Continent_Alpha2")
    second := res.Map("Continent_Name")
    for _, row := range rows {
//      alpha2 := row.Str(0)
      alpha2:=row.Str(premier)
      nom:=row.Str(second)
      fmt.Println(alpha2,"  ",nom)
    }

    fmt.Printf("------------------\n")
    pram1:="1c5bfbe4f895ba25d7869293fb81cbcb00f3996984132a3b5fef03731e9d46e0"
    fmt.Println("pram1: ",pram1)
  rows, res, err = db.Query("select DateTime,PeerId from Inputs_Peers where HashId='%s' order by DateTime asc",pram1)
  if res == nil {
    fmt.Println("no results")
  } else{
    if err != nil {
//        panic(err)
      fmt.Println("panic 01")
    }
    numrows:=len(rows)
    fmt.Println("Nombre de lignes: ",numrows)
  }
    fmt.Printf("------------------\n")
//-------------------------html server
  fmt.Println("starting web server")
  http.HandleFunc("/", rootHandler)
  http.HandleFunc("/eval", evalHandler)
  http.HandleFunc("/favicon.ico", icoHandler)
  http.ListenAndServe(":8080", nil)
}

