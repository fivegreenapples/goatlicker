package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fivegreenapples/goatlicker/datastore"
)

func main() {

	var datastoresLocation string
	flag.StringVar(&datastoresLocation, "datastores", "", "Folder path for JSON datastores")
	flag.Parse()

	if datastoresLocation == "" {
		log.Fatal("No datastores location provided")

	}

	availableDatastores := map[string]*datastore.Datastore{}

	dirEntries, dirErr := os.ReadDir(datastoresLocation)
	if dirErr != nil {
		log.Fatal(dirErr)
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.Type().IsRegular() {
			extension := filepath.Ext(dirEntry.Name())
			if strings.ToLower(extension) == ".json" {
				fp := filepath.Join(datastoresLocation, dirEntry.Name())
				ds, dsErr := datastore.Load(fp)
				if dsErr != nil {
					log.Fatal(dsErr)
				}
				log.Printf("Loaded datastore id '%s' at '%s'", ds.Identifier, fp)
				availableDatastores[ds.Identifier] = ds
			}
		}
	}

	dsFactory := func(id string) *datastore.Datastore {
		return availableDatastores[id]
	}

	allAPIHandlers := map[string]apiMethodHandler{}
	allAPIHandlers["account.get"] = account_get
	allAPIHandlers["person.getforaccount"] = person_getforaccount
	allAPIHandlers["transaction.getforaccount"] = transaction_getforaccount
	allAPIHandlers["transaction.add"] = transaction_add
	allAPIHandlers["transaction.get"] = transaction_get
	allAPIHandlers["transaction.get_payments"] = transaction_getpayments
	allAPIHandlers["transaction.update"] = transaction_add
	allAPIHandlers["transaction.delete"] = transaction_delete

	handler := makeHandler(func(module, method string, params json.RawMessage) (interface{}, error) {

		canonicalMethod := strings.ToLower(module) + "." + strings.ToLower(method)
		log.Println("API Request", canonicalMethod)

		apiHandler, found := allAPIHandlers[canonicalMethod]
		if !found {
			return nil, errors.New("unhandled method")
		}

		return apiHandler(dsFactory, params)

	})

	http.ListenAndServe("127.0.0.1:18000", handler)

}
