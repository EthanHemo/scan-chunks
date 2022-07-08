package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// var EOF = errors.New("EOF")

type fileResponse struct {
	file string
	val  byte
}

type commandType int

const (
	commandDie commandType = iota
	commandNext
	commandContinue
)

func writeOutputToFile(filename string, output []byte, buffer io.Reader) {
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	w := bufio.NewWriter(f)
	_, err = w.Write(output)
	if err != nil {
		return
	}
	io.Copy(w, buffer)
	w.Flush()
}

func readFile(baseUrl string, filename string, output chan fileResponse, commands chan commandType) {
	status := commandNext
	size := 1024

	if baseUrl[len(baseUrl)-1] != '/' {
		baseUrl = baseUrl + "/"
	}
	url := baseUrl + filename
	fmt.Printf("Download URL: %s\n", url)
	res, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	txt := make([]byte, size)
	n, err := res.Body.Read(txt)
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	fileContent := make([]byte, 0)
	for {
		for i := range txt[0:n] {
			if status == commandNext {
				output <- fileResponse{
					file: filename,
					val:  txt[i],
				}

				status = <-commands
				if status == commandDie {
					fmt.Printf("Kill download File: %s\n", filename)
					return
				}
			}
		}
		fileContent = append(fileContent, txt[0:n]...)
		if err == io.EOF {
			writeOutputToFile(filename, fileContent, res.Body)
			fmt.Printf("Finished File: %s\n", filename)
			output <- fileResponse{}
			return
		}

		n, err = res.Body.Read(txt)
		if err != nil && err != io.EOF {
			log.Fatalln(err)
		}
	}

}

func downloadFiles(url string, filelist []string) {
	output := make(chan fileResponse, len(filelist))
	mapCommands := make(map[string]chan commandType)
	detectedArr := make([]string, 0)

	for _, file := range filelist {
		mapCommands[file] = make(chan commandType)
		go readFile(url, file, output, mapCommands[file])

	}

	for {
		for i := 0; i < len(filelist); i++ {
			resp := <-output
			if resp.val == "A"[0] {
				fmt.Printf("Found A in %s\n", resp.file)
				detectedArr = append(detectedArr, resp.file)
			}
		}
		// If no A was found
		if len(detectedArr) == 0 {
			for file := range mapCommands {
				mapCommands[file] <- commandNext
			}
			// Found files will A, allow the rest of the downloads and kill the rest
		} else {
			for _, detected := range detectedArr {
				mapCommands[detected] <- commandContinue
				delete(mapCommands, detected)
			}
			for file := range mapCommands {
				mapCommands[file] <- commandDie
			}
			for i := 0; i < len(detectedArr); i++ {
				<-output
			}
			return
		}

	}
}

func getUrl() string {
	url := flag.String("url", "http://localhost:8090", "The url to connect to. i.e http://localhost:8090")
	flag.Parse()

	return *url

}

func main() {
	url := getUrl()

	res, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	sb := string(body)

	r, _ := regexp.Compile("href=\"(.+)\"")
	files := make([]string, 0)
	for _, line := range strings.Split(sb, "\n") {
		matches := r.FindStringSubmatch(line)
		if len(matches) > 0 {
			files = append(files, matches[1])
		}
	}
	downloadFiles(url, files)
}
