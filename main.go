package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"database/sql"
	"time"

    	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)


type jsonErr struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type Zeiptstr struct {
    Id    int
    RcptDate  string
    RcptData  string
    Gcid  string
}

type Receipt struct {
	Id        int       `json:"RcptID"`
	Data      string    `json:"Data"`
	//Completed bool      `json:"completed"`
	Due       time.Time `json:"RcptDate"`
}

type Receipts []Receipt

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},
	Route{
		"ReceiptIndex",
		"GET",
		"/receipts",
		ReceiptIndex,
	},
	Route{
		"ReceiptCreate",
		"POST",
		"/receipts/add",
		ReceiptCreate,
	},
	Route{
		"ReceiptShow",
		"GET",
		"/receipts/{gcId}",
		ReceiptShow,
	},
}

func main() {

	router := NewRouter()

	log.Fatal(http.ListenAndServe(":3000", router))
}



////////////////////////////////////////////////////////////////////////////

func dbConn() (db *sql.DB) {
    dbDriver := "mysql"
    dbUser := "zeiptuser"
    dbPass := "123456"
    dbName := "zeiptdb"
    db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(35.228.136.150)/"+dbName+"?charset=utf8&parseTime=True")
    if err != nil {
        panic(err.Error())
    }
    return db
}

/*func dbConn() (db *sql.DB) {
    dbDriver := "mysql"
    dbUser := "root"
    dbPass := "1234 !@#$"
    dbName := "zeiptdb"
    db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName+"?charset=utf8&parseTime=True")
    if err != nil {
        panic(err.Error())
    }
    return db
}*/

func ReceiptIndex(w http.ResponseWriter, r *http.Request) {
    db := dbConn()
    selDB, err := db.Query("SELECT * FROM ReceiptRAW ORDER BY RcptID DESC")
    if err != nil {
        panic(err.Error())
    }
    rcpt := Zeiptstr{}
    res := []Zeiptstr{}
    for selDB.Next() {
        var id int
        var gcid, rcptdata string
	var rcptdate string

        err = selDB.Scan(&id, &rcptdate ,&rcptdata ,&gcid)
        if err != nil {
            panic(err.Error())
        }

        rcpt.Id = id
	rcpt.RcptDate = rcptdate
        rcpt.RcptData = rcptdata
        rcpt.Gcid = gcid
	
        res = append(res, rcpt)
    }

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err)
	}

    defer db.Close()
}

func ReceiptShow(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    var gcid string
    
    gcid = vars["gcId"]

    db := dbConn()

    selDB, err := db.Query("SELECT * FROM ReceiptRAW WHERE GCID=?", gcid)
    if err != nil {
        panic(err.Error())
    }
    rcpt := Zeiptstr{}
    res := []Zeiptstr{}

    for selDB.Next() {
        var id int
        var gcid, rcptdata string
	var rcptdate string

        err = selDB.Scan(&id, &rcptdate ,&rcptdata ,&gcid)
        if err != nil {
            panic(err.Error())
        }

        rcpt.Id = id
	rcpt.RcptDate = rcptdate
        rcpt.RcptData = rcptdata
        rcpt.Gcid = gcid
	
        res = append(res, rcpt)
    }

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err)
	}

    defer db.Close()
}

func ReceiptCreate(w http.ResponseWriter, r *http.Request) {

    var gcid, data string
    db := dbConn()
    if r.Method == "POST" {
        gcid = r.FormValue("gcid")   
        data = r.FormValue("data")   
        insForm, err := db.Prepare("INSERT INTO ReceiptRAW(Data, GCID) VALUES(?,?)")
        if err != nil {
            panic(err.Error())
        }
        insForm.Exec(data, gcid)
        log.Println("INSERT: Data: " + data + " | GCID: " + gcid)
    }
    defer db.Close()
    http.Redirect(w, r, "/receipts", 301)
}

//////////////////////////////////////////////////////////////////////////////

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to ZEIPT Services!\n")
}


func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}

	return router
}

var currentId int

var receipts Receipts

// Give us some seed data
func init() {
	RepoCreateReceipt(Receipt{Data: "{Test One}"})
	RepoCreateReceipt(Receipt{Data: "{Test Tow}"})
}

func RepoFindReceipt(id int) Receipt {
	for _, t := range receipts {
		if t.Id == id {
			return t
		}
	}
	// return empty Todo if not found
	return Receipt{}
}

//this is bad, I don't think it passes race condtions
func RepoCreateReceipt(t Receipt) Receipt {
	currentId += 1
	t.Id = currentId
	receipts = append(receipts, t)
	return t
}

func RepoDestroyReceipt(id int) error {
	for i, t := range receipts {
		if t.Id == id {
			receipts = append(receipts[:i], receipts[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Could not find Todo with id of %d to delete", id)
}
