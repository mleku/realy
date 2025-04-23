package gui

import (
	_ "embed"
	"net/http"

	"realy.mleku.dev/servemux"
)

//go:embed index_works.html
var index []byte

//go:embed main.wasm
var wasm []byte

//go:embed wasm_works.js
var js []byte

func New(path string, sm *servemux.S) {
	sm.HandleFunc(path+"/", func(w http.ResponseWriter, req *http.Request) {
		w.Write(index)
	})
	sm.HandleFunc(path+"/main.wasm", func(w http.ResponseWriter, req *http.Request) {
		w.Write(wasm)
	})
	sm.HandleFunc(path+"/wasm.js", func(w http.ResponseWriter, req *http.Request) {
		w.Write(js)
	})
}
