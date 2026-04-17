package webserver

import (
	// "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func ServeWeb(port int) {

	webaddr := fmt.Sprintf(":%d", port)

	webfolder := http.Dir("./web")

    fs := http.FileServer(webfolder)
    // http.Handle("/", http.StripPrefix("/", fs))


    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    	// Try to serve the file normally (e.g., /js/app.js)
    	f, err := webfolder.Open(r.URL.Path[1:])
    	if err == nil {
    		f.Close()
    		fs.ServeHTTP(w, r)
    		return
    	}

    	// If file doesn't exist, serve index.html for Vue Router to take over
    	index, _ := webfolder.Open("index.html")
    	defer index.Close()
    	http.ServeContent(w, r, "index.html", time.Now(), index.(io.ReadSeeker))
    })


    log.Printf("serve web at %s", webaddr)
    http.ListenAndServe(webaddr, nil)

}
