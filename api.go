package main

import (
	"encoding/json"
	"errors"

	"github.com/fivegreenapples/goatlicker/datastore"
	"github.com/fivegreenapples/goatlicker/model"
)

type dsFactory func(string) *datastore.Datastore
type apiMethodHandler func(dsFactory, json.RawMessage) (interface{}, error)

var (
	notFoundErr      = errors.New("NOT_FOUND")
	jsonUnmarshalErr = errors.New("JSON_UNMARSHAL_ERROR")
)

func account_get(dsFactory dsFactory, params json.RawMessage) (interface{}, error) {
	reqParams := struct {
		Identifier string
	}{}
	jsonErr := json.Unmarshal(params, &reqParams)
	if jsonErr != nil {
		return nil, jsonUnmarshalErr
	}

	ds := dsFactory(reqParams.Identifier)
	if ds == nil {
		return nil, notFoundErr
	}

	return ds.GetAccount(), nil
}

func person_getforaccount(dsFactory dsFactory, params json.RawMessage) (interface{}, error) {
	reqParams := struct {
		Account_Identifier string
	}{}
	jsonErr := json.Unmarshal(params, &reqParams)
	if jsonErr != nil {
		return nil, jsonUnmarshalErr
	}

	ds := dsFactory(reqParams.Account_Identifier)
	if ds == nil {
		return nil, notFoundErr
	}

	return ds.GetPeople(), nil
}

func transaction_getforaccount(dsFactory dsFactory, params json.RawMessage) (interface{}, error) {
	reqParams := struct {
		Account_Identifier string
	}{}
	jsonErr := json.Unmarshal(params, &reqParams)
	if jsonErr != nil {
		return nil, jsonUnmarshalErr
	}

	ds := dsFactory(reqParams.Account_Identifier)
	if ds == nil {
		return nil, notFoundErr
	}

	return ds.GetTransactions(), nil
}

func transaction_add(dsFactory dsFactory, params json.RawMessage) (interface{}, error) {
	reqParams := struct {
		Account_Identifier string
		Data               struct {
			Id          int
			Description string
			Date        int
			Payments    []struct {
				Toad   int
				Amount int
			}
		}
	}{}
	jsonErr := json.Unmarshal(params, &reqParams)
	if jsonErr != nil {
		return nil, jsonUnmarshalErr
	}

	ds := dsFactory(reqParams.Account_Identifier)
	if ds == nil {
		return nil, notFoundErr
	}

	newTransaction := model.Transaction{
		Description: reqParams.Data.Description,
		Date:        reqParams.Data.Date,
	}

	transactionId := reqParams.Data.Id
	if transactionId > 0 {
		_, found := ds.UpdateTransaction(transactionId, newTransaction)
		if !found {
			return nil, notFoundErr
		}
		ds.DeletePaymentsForTransaction(transactionId)
	} else {
		newTransaction = ds.AddTransaction(newTransaction)
		transactionId = newTransaction.Id
	}

	for _, p := range reqParams.Data.Payments {
		newPayment := model.Payment{
			PersonId: p.Toad,
			Amount:   p.Amount,
		}
		ds.AddPayment(transactionId, newPayment)
	}

	return nil, nil
}

func transaction_get(dsFactory dsFactory, params json.RawMessage) (interface{}, error) {
	reqParams := struct {
		Account_Identifier string
		Id                 int
	}{}
	jsonErr := json.Unmarshal(params, &reqParams)
	if jsonErr != nil {
		return nil, jsonUnmarshalErr
	}

	ds := dsFactory(reqParams.Account_Identifier)
	if ds == nil {
		return nil, notFoundErr
	}

	return ds.GetTransactionById(reqParams.Id), nil
}

func transaction_getpayments(dsFactory dsFactory, params json.RawMessage) (interface{}, error) {
	reqParams := struct {
		Account_Identifier string
		Id                 int
	}{}
	jsonErr := json.Unmarshal(params, &reqParams)
	if jsonErr != nil {
		return nil, jsonUnmarshalErr
	}

	ds := dsFactory(reqParams.Account_Identifier)
	if ds == nil {
		return nil, notFoundErr
	}

	return ds.GetPaymentsForTransaction(reqParams.Id), nil
}

func transaction_delete(dsFactory dsFactory, params json.RawMessage) (interface{}, error) {
	reqParams := struct {
		Account_Identifier string
		Id                 int
	}{}
	jsonErr := json.Unmarshal(params, &reqParams)
	if jsonErr != nil {
		return nil, jsonUnmarshalErr
	}

	ds := dsFactory(reqParams.Account_Identifier)
	if ds == nil {
		return nil, notFoundErr
	}

	ds.DeleteTransactionById(reqParams.Id)
	return nil, nil
}
