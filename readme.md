# Scan Chunks
In this project, the application will read a file server, detect the files in the server and scan to see which one has the first A.  
After finding the files with the most recent A in their content, it will continue to download the files and kill the rest.  

## Assumptions  
The server it will connect to has only files, and will have output as following:  
``` html
<a href="file1">file1</a>
```

## Execution  
To run the application, pass the parameter `url`  
For example:  
```
scan-chunks --url http://localhost:8090
```  

## Create local file server  
To run a small golan file server use the following code:  
``` golang  
package main

import (
	"log"
	"net/http"
)

func main() {
	directory := "."
	port := "8090"
	http.Handle("/", http.FileServer(http.Dir(directory)))

	log.Printf("Serving %s on HTTP port: %s\n", directory, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
```
