package datastore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/fivegreenapples/goatlicker/model"
)

type Datastore struct {
	// Model Fields
	Identifier   string
	Account      model.Account
	People       map[int]model.Person
	Transactions map[int]model.Transaction
	Payments     map[int][]model.Payment

	// Exported internal logic
	Autoincrement int

	// Internal consistency
	filename string
	mutx     sync.RWMutex
}

func Load(filename string) (*Datastore, error) {

	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed reading db file: %s", err)
	}

	thisDB := Datastore{}
	decodeErr := json.Unmarshal(fileBytes, &thisDB)
	if decodeErr != nil {
		return nil, fmt.Errorf("Failed decode of library file: %s", decodeErr)
	}

	thisDB.filename = filename

	// todo validate consistency?

	return &thisDB, nil
}

func (ds *Datastore) save() error {

	// assumes lock is held

	bytes, jsonErr := json.MarshalIndent(ds, "", "    ")
	if jsonErr != nil {
		return fmt.Errorf("Failed marshalling db: %s", jsonErr)
	}

	writeErr := ioutil.WriteFile(ds.filename, bytes, 0644)
	if writeErr != nil {
		return fmt.Errorf("Failed saving db file: %s", writeErr)
	}

	return nil
}

func (ds *Datastore) AddPerson(p model.Person) model.Person {
	ds.mutx.Lock()
	defer ds.mutx.Unlock()

	ds.Autoincrement += 1
	p.Id = ds.Autoincrement
	ds.People[p.Id] = p
	ds.save()

	return p
}

func (ds *Datastore) AddTransaction(t model.Transaction) model.Transaction {
	ds.mutx.Lock()
	defer ds.mutx.Unlock()

	ds.Autoincrement += 1
	t.Id = ds.Autoincrement
	ds.Transactions[t.Id] = t
	ds.save()

	return t
}

func (ds *Datastore) UpdateTransaction(transactionId int, t model.Transaction) (model.Transaction, bool) {
	ds.mutx.Lock()
	defer ds.mutx.Unlock()

	currentT, found := ds.Transactions[transactionId]
	if !found {
		return model.Transaction{}, false
	}

	currentT.Date = t.Date
	currentT.Description = t.Description
	ds.Transactions[transactionId] = currentT
	ds.save()

	return currentT, true
}

func (ds *Datastore) AddPayment(transactionId int, p model.Payment) model.Payment {
	ds.mutx.Lock()
	defer ds.mutx.Unlock()

	if ds.Payments == nil {
		ds.Payments = map[int][]model.Payment{}
	}
	ds.Payments[transactionId] = append(ds.Payments[transactionId], p)

	// Adjust balance for person
	person := ds.People[p.PersonId]
	person.Balance += p.Amount
	ds.People[p.PersonId] = person

	// Calc total amount for this transaction
	totalAmount := 0
	for _, payment := range ds.Payments[transactionId] {
		if payment.Amount > 0 {
			totalAmount += payment.Amount
		}
	}
	t := ds.Transactions[transactionId]
	t.TotalAmount = totalAmount
	ds.Transactions[transactionId] = t

	ds.save()

	return p
}

func (ds *Datastore) GetAccount() model.Account {
	ds.mutx.RLock()
	defer ds.mutx.RUnlock()

	return ds.Account
}

func (ds *Datastore) GetPeople() []model.Person {
	ds.mutx.RLock()
	defer ds.mutx.RUnlock()

	people := []model.Person{}
	for _, p := range ds.People {
		people = append(people, p)
	}
	return people
}

func (ds *Datastore) GetTransactions() []model.Transaction {
	ds.mutx.RLock()
	defer ds.mutx.RUnlock()

	transactions := []model.Transaction{}
	for _, t := range ds.Transactions {
		transactions = append(transactions, t)
	}
	return transactions
}

func (ds *Datastore) GetTransactionById(transactionId int) model.Transaction {
	ds.mutx.RLock()
	defer ds.mutx.RUnlock()

	return ds.Transactions[transactionId]
}

func (ds *Datastore) DeleteTransactionById(transactionId int) {
	ds.mutx.Lock()
	defer ds.mutx.Unlock()

	ds.deletePaymentsAndAdjustBalances(transactionId)
	delete(ds.Transactions, transactionId)
	ds.save()
}

func (ds *Datastore) GetPaymentsForTransaction(transactionId int) []model.Payment {
	ds.mutx.RLock()
	defer ds.mutx.RUnlock()

	if ds.Payments == nil {
		return []model.Payment{}
	}

	payments, found := ds.Payments[transactionId]
	if !found {
		return []model.Payment{}
	}

	for idx, p := range payments {
		payments[idx].Name = ds.People[p.PersonId].Name
	}
	return payments
}

func (ds *Datastore) DeletePaymentsForTransaction(transactionId int) {
	ds.mutx.Lock()
	defer ds.mutx.Unlock()

	ds.deletePaymentsAndAdjustBalances(transactionId)
	ds.save()
}

func (ds *Datastore) deletePaymentsAndAdjustBalances(transactionId int) {
	// assumes lock is held

	for _, p := range ds.Payments[transactionId] {
		person := ds.People[p.PersonId]
		person.Balance -= p.Amount
		ds.People[p.PersonId] = person
	}
	delete(ds.Payments, transactionId)
}
