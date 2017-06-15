package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type serverInfo struct {
	addr     string
	port     string
	password string
	method   string
}

// NewServerInfo creates a serverInfo instance
// It returns a pointer to serverInfo
func newServerInfo(addr, port, password, method string) *serverInfo {
	return &serverInfo{addr, port, password, method}
}

func getSSInfo(url string) (srvs []*serverInfo) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatalln(err)
	}
	srvs = make([]*serverInfo, 0)
	doc.Find("div.col-md-6").Each(func(i int, s *goquery.Selection) {
		var addr, port, password, method string
		s.Find("h4").Each(func(j int, sel *goquery.Selection) {
			text := sel.Text()
			if pos := strings.Index(text, ":"); pos > -1 {
				switch {
				case strings.Contains(text, "服务器地址:"):
					addr = text[pos+1:]
				case strings.Contains(text, "端口:"):
					port = text[pos+1:]
				case strings.Contains(text, "密码:"):
					password = text[pos+1:]
				case strings.Contains(text, "加密方式:"):
					method = text[pos+1:]
				}
			}
		})
		if addr != "" && port != "" && method != "" {
			server := newServerInfo(addr, port, password, method)
			srvs = append(srvs, server)
		}
	})

	return
}

func writeSSInfo(filePath string, srvs *[]*serverInfo) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	var obj interface{}
	if err = json.Unmarshal(data, &obj); err != nil {
		log.Fatalln(err)
	}
	m := obj.(map[string]interface{})
	/*
		configs := m["configs"].([]interface{})
		for _, v := range configs {
			server := v.(map[string]interface{})
			server["password"] = "golang"
			fmt.Println(server)
		}
	*/
	servers := make([]map[string]interface{}, 0)
	for _, val := range *srvs {
		server := make(map[string]interface{})
		server["server"] = val.addr
		i, err := strconv.Atoi(val.port)
		if err != nil {
			log.Fatalln(err)
		}
		server["server_port"] = i
		server["password"] = val.password
		server["method"] = val.method
		server["remarks"] = ""
		server["auth"] = false
		server["timeout"] = 5
		servers = append(servers, server)
	}
	m["configs"] = servers
	//fmt.Println(obj)
	bytes, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		log.Fatalln(err)
	}
	err = ioutil.WriteFile(filePath, bytes, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func startSS(filePath string) {
	cmd := exec.Command(filePath)
	err := cmd.Start()
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	if runtime.GOOS != "windows" {
		log.Fatalln("This is currently for Windows only :)")
		return
	}
	url := "https://freessr.xyz/"
	servers := getSSInfo(url)

	cfgFile := "./gui-config.json"
	writeSSInfo(cfgFile, &servers)

	exeFile := "./Shadowsocks.exe"
	startSS(exeFile)
}
