// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// https://github.com/gorilla/websocket/blob/master/LICENSE

package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "../mahjong-ui/mahjong-ui.html")
}

func main() {
	flag.Parse()
	m := MahjongPlayManager{}
	m.Init()
	hub := newHub(&m)
	go hub.run()
	http.HandleFunc("/", serveHome)
	http.Handle("/mahjong-ui/", http.StripPrefix("/mahjong-ui/", http.FileServer(http.Dir("../mahjong-ui"))))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
