package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	pathP "path"
	"regexp"
	"runtime"
	"text/template"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	name     = 12032
	pwd      = "supersecret"
	platform string
	path     = ""
	port     string
	fileList []*File
)

// File type
type File struct {
	Name   string `json:"name"`
	Folder bool   `json:"folder"`
	Image  bool   `json:"image"`
	Video  bool   `json:"video"`
	Size   int    `json:"size"`
}

func authServer(cb func()) {
	var (
		uName int
		uPwd string
	)
	fmt.Print("n: ")
	fmt.Scan(&uName)
	fmt.Print("w: ")
	if platform == "linux" || platform == "darwin" {
		tuPwd, err := terminal.ReadPassword(0)
		if err != nil {
			fmt.Println(err)
			return
		}
		uPwd = string(tuPwd)
	} else {
		fmt.Scan(&uPwd)
	}
	
	if !(name == uName && pwd == uPwd) {
		log.Fatal("[failed]")
		return
	}
	cb()
}

func setupStatic() {
	fmt.Print("[path]:")
	fmt.Scan(&path)
	fmt.Print("PORT:")
	fmt.Scan(&port)
	http.Handle("/", http.FileServer(http.Dir(path)))
}

func serverFile(w http.ResponseWriter, r *http.Request) {

	fp := pathP.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, map[string]string{
		"path": path,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createFile(f fs.FileInfo) *File {
	var file File
	file.Folder = f.IsDir()
	file.Size = int(f.Size())
	file.Name = f.Name()
	file.Image = false
	file.Video = false
	isImage, err := regexp.MatchString("{.jpeg|.jpg|.png|.webp}", f.Name())
	if err != nil {
		fmt.Println(err)
	} else {
		if isImage {
			file.Image = true
		}
	}

	isVideo, err := regexp.MatchString(".mp4", f.Name())
	if err != nil {
		fmt.Println(err)
	} else {
		if isVideo {
			file.Video = true
		}
	}
	return &file
}

func populateItems(p string) {
	fmt.Println(p)
	fileList = []*File{}
	files, err := ioutil.ReadDir(p)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, v := range files {
		fileList = append(fileList, createFile(v))
	}
	fmt.Println(fileList)
}

func openCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")
}

func getData(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	// openCors(w)
	param := r.URL.Query()["path"]
	fmt.Println(param)
	if len(param) == 0 {
		w.Write([]byte("no1"))
		return
	}
	value := path + param[0]
	fmt.Println(value)
	populateItems(value)
	data, err := json.Marshal(&fileList)
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte("no2"))
		return
	}
	w.Write(data)
	return
}

func main() {
	platform = runtime.GOOS
	populateItems("/Users/arunv/Documents/projects")
	authServer(func() {
		http.HandleFunc("/dash", serverFile)
		http.HandleFunc("/api", getData)
		setupStatic()
		fmt.Println("server started")
		http.ListenAndServe(port, nil)
	})
}
