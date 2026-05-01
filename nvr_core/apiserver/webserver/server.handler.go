package webserver

import (
	// "fmt"
	// "io"
	// "log"
	"net/http"
	// "path/filepath"
	// "time"
	"nvr_core/utils"
)

func ServeWeb(mux *http.ServeMux,config *utils.Config) {

	webfolder := http.Dir("./web")
	fs := http.FileServer(webfolder)

	// Serve config.js
	mux.HandleFunc("/config.js", HandleGenerateConfig(GenerateConfig(config)))

	mux.Handle("/", fs)


	// SPA Route & Static Assets
	// mux.HandleFunc("/web/", func(w http.ResponseWriter, r *http.Request) {

	// 	// ALWAYS use the raw r.URL.Path. Stripping the slash [1:] can break 
	// 	// http.Dir's internal absolute path resolution on some operating systems.
	// 	f, err := webfolder.Open(r.URL.Path)
	// 	if err == nil {
	// 		defer f.Close()
	// 		stat, err := f.Stat()
			
	// 		// Only serve if it's a real file. If it's a directory (like /assets/), 
	// 		// let it fall through rather than exposing directory contents.
	// 		if err == nil && !stat.IsDir() {
	// 			fs.ServeHTTP(w, r)
	// 			return
	// 		}
	// 	}

	// 	// --- Assets Query Check ---
	// 	// If the browser is asking for a specific asset file (.js, .css, .png, etc.) 
	// 	// and it was NOT found on disk, we MUST return a 404. 
	// 	ext := filepath.Ext(r.URL.Path)
	// 	if ext != "" && ext != ".html" {
	// 		http.NotFound(w, r)
	// 		return
	// 	}

	// 	// --- SERVE INDEX.HTML ---
	// 	log.Printf("Serving Vue Router fallback for: %s", r.URL.Path)
	// 	index, err := webfolder.Open("/index.html")
	// 	if err != nil {
	// 		// If index is nil, calling index.Close() or index.(io.ReadSeeker) will PANIC and crash Go.
	// 		// We catch the error safely here.
	// 		http.Error(w, "index.html missing from ./web folder", http.StatusNotFound)
	// 		return
	// 	}
	// 	defer index.Close()

	// 	http.ServeContent(w, r, "index.html", time.Now(), index.(io.ReadSeeker))
	// })

	// log.Printf("[Web Server] Starting at %s", webaddr)
	// if err := http.ListenAndServe(webaddr, nil); err != nil {
	// 	log.Fatalf("Web server failed: %v", err)
	// }
}