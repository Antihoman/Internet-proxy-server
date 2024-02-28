
package repository

import (
	"context"
	"log"

	"github/Antihoman/Internet-proxy-server/cmd/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	db            *mongo.Client
	reqCollection *mongo.Collection
}

func NewRepository(db *mongo.Client) (Repository, error) {
	coll := db.Database("mitm").Collection("requests")
	_, err := coll.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{"host": 1},
	})
	if err != nil {
		log.Println("error Indexes CreateOne", err)
		return Repository{}, err
	}
	return Repository{db: db, reqCollection: coll}, nil
}

func (repo *Repository) Add(req domain.Request) error {
	_, err := repo.reqCollection.InsertOne(context.TODO(), req)
	if err != nil {
		log.Println("error insert one", err)
		return err
	}
	return nil
}

func (repo *Repository) GetAll() ([]domain.Request, error) {
	cur, err := repo.reqCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Println("error Find", err)
		return []domain.Request{}, err
	}
	defer cur.Close(context.TODO())

	var results []domain.Request
	for cur.Next(context.TODO()) {
		var result domain.Request
		err := cur.Decode(&result)
		if err != nil {
			log.Println("error Decode", err)
			continue
		}
		results = append(results, result)
	}
	if err := cur.Err(); err != nil {
		log.Println("error cur.Err", err)
		return []domain.Request{}, err
	}

	return results, nil
}