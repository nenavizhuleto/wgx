package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/nenavizhuleto/wgx"
)

func main() {
	wg, err := wgx.NewServer()
	if err != nil {
		panic(err)
	}

	if err := wg.Sync(); err != nil {
		panic(err)
	}

	http.HandleFunc("POST /client", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		var req struct {
			Name string `json:"name"`
		}

		if err := json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		client, err := wg.AddClient(req.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		var res struct {
			ID         string `json:"id"`
			Address    string `json:"address"`
			PrivateKey string `json:"private_key"`
			PublicKey  string `json:"public_key"`
		}

		res.ID = client.ID.String()
		res.Address = client.Address
		res.PublicKey = string(wg.PublicKey)
		res.PrivateKey = string(client.PrivateKey)

		payload, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(payload)

		if err := wg.Sync(); err != nil {
			panic(err)
		}

		return
	})

	if err := http.ListenAndServe(":51821", nil); err != nil {
		panic(err)
	}
}
