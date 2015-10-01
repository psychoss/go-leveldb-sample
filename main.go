package main

import (
	"encoding/xml"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Explain struct {
	EX []string `xml:"ex"`
}

type Basic struct {
	Phonetic string  `xml:"phonetic"`
	Explain  Explain `xml:"explains"`
}

type Result struct {
	//Base  xml.Name `xml:"youdao-fanyi"`
	Query string `xml:"query"`
	Basic Basic  `xml:"basic"`
}

var (
	db  *leveldb.DB
	err error
)

const (
	YOUDAOAPI = "http://fanyi.youdao.com/openapi.do?keyfrom=assadsad&key=377354510&type=data&doctype=xml&version=1.1&q="
)

func main() {
	db, err = leveldb.OpenFile("db", nil)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", handleRequest)
	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		panic(err)
	}
}

func fromLevel(query string) ([]byte, error) {
	data, err := db.Get([]byte(query), nil)
	return data, err
}

func fromYouDao(query string, w http.ResponseWriter) {
	res, err := http.Get(YOUDAOAPI + url.QueryEscape(query))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	v := Result{}
	xml.Unmarshal(data, &v)
	pretty(query, v, w)
}

//Sometimes 'query' is not equal the youdao api return
//so we insert the user input 'query' as leveldb key
func pretty(queryUser string, v Result, w http.ResponseWriter) {
	query := v.Query
	content := " /" + v.Basic.Phonetic + "/ "
	for _, i := range v.Basic.Explain.EX {

		content += i + "; "
	}
	out := []byte(query + "\n" + content)
	db.Put([]byte(queryUser), out, nil)
	w.Write(out)
	fmt.Println(content)
}
func handleRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		http.ServeFile(w, r, "")
	} else {
		query := strings.Trim(r.FormValue("query"), "\r\n\t ")
		data, err := fromLevel(query)
		if err != nil {
			fromYouDao(query, w)
		}
		w.Write(data)
		fmt.Println(data, err)

	}
	fmt.Println(r.URL.Path)

}
