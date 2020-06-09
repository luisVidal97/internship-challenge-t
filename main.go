package main

//---------------------------------------------------------------------------------------------------------
//	Imprtations
//---------------------------------------------------------------------------------------------------------
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/buaazp/fasthttprouter"
	"github.com/likexian/whois-go"
	"github.com/valyala/fasthttp"
)

//---------------------------------------------------------------------------------------------------------
// Vairables an structures
//---------------------------------------------------------------------------------------------------------

var globalDB *sql.DB

// DataDomain is...
type DataDomain struct {
	Servers          []DataServer
	ServersChange    bool
	SSLGrade         string
	PreviousSSLGrade string
	Logo             string
	Title            string
	IsDown           bool
}

// DataServer is...
type DataServer struct {
	Address  string
	SSLGrade string
	Country  string
	Owner    string
}

//---------------------------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------------------------

// Index is ...
func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "Welcome!\n")
}

// Hello ...
func Hello(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}

// Post ...
func Post(ctx *fasthttp.RequestCtx) {

	var domain interface{} = ctx.UserValue("domain")
	err := registerDomain(globalDB, domain.(string))

	ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))

	if err != nil {
		ctx.Response.SetStatusCode(500)
		fmt.Println(err)
		json.NewEncoder(ctx).Encode("Error interno del servidor")
	} else {
		ctx.Response.SetStatusCode(201)
		json.NewEncoder(ctx).Encode("Se ha registrado con éxito")
	}

}

// GetDomains ...
func GetDomains(ctx *fasthttp.RequestCtx) {

	data, err := ConsultDomains(globalDB)

	ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))

	if err != nil {
		ctx.Response.SetStatusCode(500)
		fmt.Println(err)
		json.NewEncoder(ctx).Encode("Error interno del servidor")
	} else {

		ctx.Response.SetStatusCode(200)
		json.NewEncoder(ctx).Encode(data)
	}
}

// save and run project atthe same time: CompileDaemon -command="internship-challenge-t.exe"
//---------------------------------------------------------------------------------------------------------
// Main method that execute whole program
//---------------------------------------------------------------------------------------------------------
func main() {

	db, err := Conn()

	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	} else {
		globalDB = db
		router := fasthttprouter.New()
		router.GET("/", Index)
		router.GET("/hello/:name", Hello)
		router.POST("/post/:domain", Post)
		router.GET("/consult", GetDomains)

		router.POST("/checkDomain/:domain", CheckDomain)

		log.Fatal(fasthttp.ListenAndServe(":8082", router.Handler))

	}
}

// CheckDomain ...
//---------------------------------------------------------------------------------------------------------
//
//---------------------------------------------------------------------------------------------------------
func CheckDomain(ctx *fasthttp.RequestCtx) {

	ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))
	var do interface{} = ctx.UserValue("domain")
	domain := do.(string)

	// get SSL and WHOIS information
	data := ObtainDataDomain(domain)
	if data == nil {
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode("The domain " + domain + " is not found")
	}

	//Store domain in database
	err := SaveDomain(domain)

	if err != nil {
		ctx.Response.SetStatusCode(500)
		fmt.Println(err)
		json.NewEncoder(ctx).Encode("Error interno del servidor")
	} else {

		ctx.Response.SetStatusCode(200)
		json.NewEncoder(ctx).Encode(data)
	}

}

// ObtainDataDomain ...
//---------------------------------------------------------------------------------------------------------
//
//---------------------------------------------------------------------------------------------------------
func ObtainDataDomain(domain string) *DataDomain {

	information := new(DataDomain)
	res, err := http.Get("https://api.ssllabs.com/api/v3/analyze?host=" + domain)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var data map[string]interface{}
	error := json.Unmarshal([]byte(string(body)), &data)
	if error != nil {
		panic(error)
	}

	test := data["status"]
	if test == "READY" {
		information.IsDown = false
	} else {
		information.IsDown = true
	}

	messages := data["endpoints"].([]interface{})

	for _, item := range messages {
		var ds DataServer
		messages2 := item.(map[string]interface{})
		//fmt.Println(messages2["ipAddress"])
		ds.Address = messages2["ipAddress"].(string)
		ds.SSLGrade = messages2["grade"].(string)
		result, err := whois.Whois(messages2["ipAddress"].(string))

		fmt.Println(result)

		if err == nil {

			res := strings.Contains(result, "No match for")
			if !res {

				orga := GetStringInBetween(result, "Organization:", "RegDate:")
				if orga != "" {
					ds.Owner = orga
				} else {
					ds.Owner = "Not found"
				}
				//fmt.Println(result)
				//result2 := GetStringInBetween(result, "PostalCode:", "Comment:")
				count := GetStringInBetween(result, "Country:", "RegDate:")
				if count != "" {
					ds.Country = count
				} else {
					ds.Country = "Not found"
				}

			} else {
				ds.Country = "Not found"
				ds.Owner = "Not found"
			}

		}

		information.AddItem(ds)
	}
	information.Title = GetTitlePage(domain)
	information.Logo = logo
	return information
}

// AddItem ...
func (dd *DataDomain) AddItem(item DataServer) []DataServer {
	dd.Servers = append(dd.Servers, item)
	return dd.Servers
}

//https://stackoverflow.com/questions/26916952/go-retrieve-a-string-from-between-two-characters-or-other-strings
// GetStringInBetween Returns empty string if no start string found
// func GetStringInBetween(str string, start string, end string) (result string) {
// 	s := strings.Index(str, start)
// 	if s == -1 {
// 		return
// 	}
// 	s += len(start)
// 	e := strings.Index(str, end)
// 	if e == -1 {
// 		return
// 	}
// 	strTrimSpace := strings.TrimSpace(str[s:e])
// 	return strTrimSpace
// }

func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	newStr := str[s:len(str)]
	e := strings.Index(newStr, end)
	if e == -1 {
		return
	}
	strTrimSpace := strings.TrimSpace(newStr[0:e])
	return strTrimSpace
}

// ObtainDataDomain ...
//---------------------------------------------------------------------------------------------------------
//
//---------------------------------------------------------------------------------------------------------
func SaveDomain(domain string) error {

	err := registerDomain(globalDB, domain)
	if err != nil {
		return err
	}
	return nil

}
