package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type apiHandler func(module string, method string, params json.RawMessage) (interface{}, error)

func makeHandler(apihandler apiHandler) http.Handler {
	handlePOST := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Access-Control-Allow-Origin", "*")

		type apiReq struct {
			Module string          `json:"module,omitempty"`
			Method string          `json:"method,omitempty"`
			Params json.RawMessage `json:"params,omitempty"`
		}

		bodyBytes, bodyErr := ioutil.ReadAll(r.Body)
		if bodyErr != nil {
			log.Printf("POST req - failed reading body: %s", bodyErr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		thisReq := apiReq{}
		jsonErr := json.Unmarshal(bodyBytes, &thisReq)
		if jsonErr != nil {
			log.Printf("POST req - failed unmarshall: %s", jsonErr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result, apiErr := apihandler(thisReq.Module, thisReq.Method, thisReq.Params)
		if apiErr != nil {
			log.Printf("POST req - failed api call: %s", apiErr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		thisResp := map[string]interface{}{
			"result": map[string]interface{}{
				"api":    "OK",
				"action": "OK",
			},
			"response": result,
		}

		resultBytes, jsonErr := json.MarshalIndent(thisResp, "", "    ")
		if jsonErr != nil {
			log.Printf("POST req - failed marshaling api result: %s", jsonErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(resultBytes)
	}

	handleOPTIONS := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST,OPTIONS")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlePOST(w, r)
			return
		}
		if r.Method == http.MethodOptions {
			handleOPTIONS(w, r)
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}
