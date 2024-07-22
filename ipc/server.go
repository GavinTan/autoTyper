package ipc

import (
	"encoding/json"
	"net/http"

	"fyne.io/fyne/v2"
)

func NewServer(fw fyne.Window) {
	DestroyConn()

	ln, err := Listen()
	if err == nil {
		m := http.NewServeMux()
		m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		})

		m.HandleFunc(PingPath, func(w http.ResponseWriter, r *http.Request) {
			var res Response
			b, _ := json.Marshal(&res)
			w.Write(b)
		})

		m.HandleFunc(ShowPath, func(w http.ResponseWriter, r *http.Request) {
			fw.Show()
		})

		http.Serve(ln, m)
	}
}
