package router

import (
	"net/http"
	"strconv"
)

func GetPackage() func(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) error {
		goGetParam := r.URL.Query().Get("go-get")
		goGet, err := strconv.Atoi(goGetParam)
		if err != nil {
			return err
		}

		if goGet != 1 {
		}

		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			//TODO
		}
	}
}
