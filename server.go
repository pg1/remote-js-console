/**
 * Remote javascript console
 * - Collect remote logs
 * - Auth and serve admin request
 */

package main

import (
	"net/http"
	"os"
	"encoding/json"
	"io"
	"fmt"
	"strconv"
	"log"
	"net"
	"strings"
	"html/template"
	"database/sql"
	"github.com/gorilla/mux"
	"encoding/base64"
	_ "github.com/go-sql-driver/mysql"

)

var db *sql.DB
type Log struct {
	Ip string	
	Timestamp string
	Useragent string
	Log string
}

type Config struct {
	Database map[string]string
	Admin map[string]string
	Server map[string]string
}
var conf Config

func logHandler(w http.ResponseWriter, r *http.Request) {

	//save log to db
	if r.Method == "GET" {
		ip,_,_ := net.SplitHostPort(r.RemoteAddr)
		msg := r.URL.Query().Get("msg")
		ua := r.UserAgent()
		data := Log{Ip:ip, Log:msg, Useragent:ua}
		go addLog(data)
	}

	//send 1x1 px gif
	w.Header().Set("Content-Type","image/gif")
	output,_ := base64.StdEncoding.DecodeString("R0lGODlhAQABAJAAAP8AAAAAACH5BAUQAAAALAAAAAABAAEAAAICBAEAOw==")
	io.WriteString(w, string(output))
}

func addLog(l Log){
	_, err := db.Exec("INSERT INTO remotelog(ip, useragent, log) VALUES(?, ?, ?)", l.Ip, l.Useragent, l.Log)
	if err != nil {
		log.Println("Warning: Write to db failed: ", err)
	}
}

func getLogs(f string)[]Log{

	//create sql and get logs
	sql := "SELECT ip, log, useragent, tmstamp FROM remotelog"

	if(len(f)>0){
		fq := strconv.Quote(f)
		sql = fmt.Sprintf("%s WHERE log LIKE %s OR ip LIKE %s OR useragent LIKE %s ", sql, fq, fq, fq)
	}

	sql = fmt.Sprintf("%s ORDER BY id DESC LIMIT 2000", sql)

	//get last 2000 rows
	rows, err := db.Query(sql)
	if err != nil {
		log.Println(err.Error())
	}

	//create dataset
	result := make([]Log, 0)
	for rows.Next() {
		var ip, log, useragent, tmstamp string
		rows.Scan(&ip, &log, &useragent, &tmstamp)
		result = append(result, Log{ip, tmstamp, useragent, log})
	}

	return result
}

func checkAuth(w http.ResponseWriter, r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 { 
		return false 
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil { 
		return false 
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 { 
		return false 
	}

	return pair[0] == conf.Admin["User"] && pair[1] == conf.Admin["Password"]
}


func adminHandler(w http.ResponseWriter, r *http.Request) {

	//check if user logged in
	if checkAuth(w, r) {
		filter := ""
		if r.Method == "GET" {
			filter = r.URL.Query().Get("filter")
		}
		logList := getLogs(filter)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t, _ := template.ParseFiles("admin/index.html") 
		t.Execute(w, map[string]interface{}{"Rows":logList, "Filter":filter})
	}else{
		w.Header().Set("WWW-Authenticate", `Basic realm="Authenticate"`)
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
	}
	
}

func main(){
	var err error

	//read config
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatalln("Config error: ", err.Error())
	}

	//connect to db    
	db, err = sql.Open("mysql", conf.Database["User"]+":"+conf.Database["Password"]+"@tcp("+conf.Database["Host"]+":"+conf.Database["Port"]+")/"+conf.Database["Dbname"])
	defer db.Close()

	//check for db error and create table if not exists
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS remotelog (id INT NOT NULL AUTO_INCREMENT, ip VARCHAR(30) NULL, log  VARCHAR(500)  NULL, useragent VARCHAR(200) NULL, tmstamp TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (id))");
	if err != nil {
		log.Fatalln("Database error: ", err.Error())
		return
	}


	//setup routing
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", logHandler)
	r.HandleFunc("/admin/", adminHandler)

	http.Handle("/", r)

	log.Println("Starting server at " + conf.Server["Url"] + ":" + conf.Server["Port"] + " ...")

	err = http.ListenAndServe(conf.Server["Url"] + ":" + conf.Server["Port"], nil)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
