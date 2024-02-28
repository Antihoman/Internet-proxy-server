package mongoclient

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const pingTimeout = 3 * time.Second

func NewMongoClient(uri string) (*mongo.Client, func(), error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, nil, err
	}

	close := func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}
	return client, close, nil
}