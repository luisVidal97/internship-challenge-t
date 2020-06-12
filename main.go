package main

//---------------------------------------------------------------------------------------------------------
//	Importations
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
// Vairables and structures
//---------------------------------------------------------------------------------------------------------

// Global variable of database in Cockroach
var globalDB *sql.DB

// DataDomain is...
type DataDomain struct {
	Servers          []DataServer
	domain           string
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

// StructSend is...
type StructSend struct {
	ID               string
	Domain           string
	PreviousSslGrade string
	CheckedAt        string
}

//---------------------------------------------------------------------------------------------------------
//----------------------------------- FUNCTIONS -----------------------------------------------------------
//---------------------------------------------------------------------------------------------------------

// GetDomains is a function that return all registries of domains checked in the app. This return a list using
// the StructSend structure.
func GetDomains(ctx *fasthttp.RequestCtx) {

	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))

	register, err := ConsultDomains(globalDB)

	if err != nil {
		ctx.Response.SetStatusCode(500)
		fmt.Println(err)
		json.NewEncoder(ctx).Encode("Error interno del servidor")
	} else {
		ctx.Response.SetStatusCode(200)
		json.NewEncoder(ctx).Encode(register)
	}
}

// CheckDomain is a function that check a domain using Whois and ssl APIS.
func CheckDomain(ctx *fasthttp.RequestCtx) {

	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))

	//type assertions
	// var do interface{} = ctx.UserValue("domain")
	// domain := do.(string)
	domain := ctx.UserValue("domain").(string)

	res, err := http.Get("https://api.ssllabs.com/api/v3/analyze?host=" + domain)

	if err != nil {

		log.Fatalln(err)
		ctx.Response.SetStatusCode(500)
		json.NewEncoder(ctx).Encode("Error when consulting 'https://api.ssllabs.com/api/v3/analyze?host'. Try again later")

	} else {

		defer res.Body.Close()

		// This return a list of bytes
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalln(err)
		}

		var data map[string]interface{}
		error := json.Unmarshal(body, &data)
		if error != nil {
			panic(error)
		}

		status := data["status"]
		if status == "READY" {

			// get SSL and WHOIS information
			data := ObtainDataDomain(data)

			//get title of page
			data.Title = GetTitlePage(domain)
			data.Logo = logo
			data.domain = domain

			//Store domain in database
			error := RegisterDom(globalDB, data)

			if error != nil {
				ctx.Response.SetStatusCode(500)
				fmt.Println(error)
				json.NewEncoder(ctx).Encode("Error interno del servidor")
			} else {

				ctx.Response.SetStatusCode(200)
				json.NewEncoder(ctx).Encode(data)
			}

		} else if status == "ERROR" {
			fmt.Println(data["statusMessage"])
			ctx.Response.SetStatusCode(404)
			json.NewEncoder(ctx).Encode("Domain not found")
		} else {
			fmt.Println(data["statusMessage"])
			ctx.Response.SetStatusCode(500)
			json.NewEncoder(ctx).Encode("Error.. " + data["statusMessage"].(string))
		}

	}

}

// ObtainDataDomain is a function that return a DataDomain struct that contain the information
// about the domain, servers, etc
func ObtainDataDomain(data map[string]interface{}) *DataDomain {

	information := new(DataDomain)
	information.IsDown = true
	// These items are in progress
	information.PreviousSSLGrade = "In progress..."
	information.ServersChange = false

	messages := data["endpoints"].([]interface{})

	for _, item := range messages {
		var ds DataServer
		messages2 := item.(map[string]interface{})
		ds.Address = messages2["ipAddress"].(string)
		ds.SSLGrade = messages2["grade"].(string)
		result, err := whois.Whois(messages2["ipAddress"].(string))

		if err == nil {

			res := strings.Contains(result, "No match for")
			if !res {

				orga := GetStringInBetween(result, "Organization:", "RegDate:")
				if orga != "" {
					ds.Owner = orga
				} else {
					ds.Owner = "Not found"
				}

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

	information.SSLGrade = GradeSmallest(information.Servers)
	return information
}

// GradeSmallest is a function that return the smallest grade of SSL certificate
func GradeSmallest(array []DataServer) string {

	data := array[0].SSLGrade
	for i := 0; i < len(array); i++ {
		fmt.Println(int(array[i].SSLGrade[0]))
		if i+1 != len(array) {
			if int(array[i+1].SSLGrade[0]) > int(array[i].SSLGrade[0]) {
				data = array[i+1].SSLGrade
			} else if int(array[i+1].SSLGrade[0]) == 65 && int(array[i].SSLGrade[0]) == 65 {
				// A+ or A or A-
				if len(array[i+1].SSLGrade) > 1 {
					if int(array[i+1].SSLGrade[1]) == 45 {
						data = array[i+1].SSLGrade
					}
				}
			}
		}
	}
	return data
}

// AddItem is a function to add item with information of Server to his domain.
func (dd *DataDomain) AddItem(item DataServer) []DataServer {
	dd.Servers = append(dd.Servers, item)
	return dd.Servers
}

// GetStringInBetween is a function used when I do a request to Whois and the response is
// a text without format. I tried to use WhoisParse but it doesn´t function very well. So,
// the idea is extract the two field: organization and country, located between words.
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

// Save and run project atthe same time: CompileDaemon -command="internship-challenge-t.exe"
//---------------------------------------------------------------------------------------------------------
// Main method that execute whole program.
//---------------------------------------------------------------------------------------------------------
func main() {

	db, err := Conn()

	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	} else {
		globalDB = db
		errorCreate := CreateTable()
		if errorCreate == nil {

			router := fasthttprouter.New()
			router.GET("/consultDomains", GetDomains)
			router.POST("/checkDomain/:domain", CheckDomain)

			log.Fatal(fasthttp.ListenAndServe(":8082", router.Handler))

		} else {
			log.Fatal("Error to creating table: ", errorCreate)
		}
	}
}
