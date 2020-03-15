package db

import (
	"context"
	"errors"
	"fmt"
	_ "github.com/go-kivik/couchdb"
	"github.com/go-kivik/kivik"
	"github.com/google/uuid"
)

type DB struct {
	*kivik.DB
}

func NewClient(driverName, dataSourceName, dbName string) (DB, error) {
	client, err := kivik.New(driverName, dataSourceName)
	if err != nil {
		return DB{}, err
	}
	database := client.DB(context.TODO(), dbName)
	return DB{database}, database.Err()
}

type document struct {
	ID  string `json:"_id,omitempty"`
	Rev string `json:"_rev,omitempty"`
}

type dBDoc struct {
	Name string `json:"name,omitempty"`
}

func (db DB) GetDocument(ctx context.Context, docID string) (document, error) {
	row := db.Get(ctx, docID)
	var doc document
	err := row.ScanDoc(&doc)
	if err != nil {
		return document{}, errors.New("this is unexpected")
	}
	return doc, nil

}

func (db DB) PutDocument(ctx context.Context) (flag bool, err error) {
	id := uuid.New().String()
	doc := dBDoc{
		Name: "hello",
	}
	_, err = db.Put(ctx, id, doc)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return true, nil
}
