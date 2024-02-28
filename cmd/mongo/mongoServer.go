package mongo

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const uri = "mongodb://root:root@localhost:27017"

type Request struct {
	ID      string            `bson:"_id,omitempty"`
	Host    string            `bson:"host"`
	Method  string            `bson:"method"`
	Version string            `bson:"version"`
	Path    string            `bson:"path"`
	Headers map[string]string `bson:"headers, omitempty"`
	Time    time.Time         `bson:"time"`
}

func main() {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	collection := client.Database("mitm").Collection("requests")
	_, err = collection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{"host": 1},
	})
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		headers := make(map[string]string)
		for name, values := range r.Header {
			for _, value := range values {
				headers[name] = value
			}
		}

		req := Request{
			Host:    r.Host,
			Method:  r.Method,
			Version: r.Proto,
			Path:    r.URL.Path,
			Headers: headers,
			Time:    time.Now(),
		}

		_, err := collection.InsertOne(context.TODO(), req)
		if err != nil {
			log.Println("error insert one")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Op! New request is saved in DB"))
	})

	http.HandleFunc("/requests", func(w http.ResponseWriter, r *http.Request) {
		cur, err := collection.Find(context.TODO(), bson.D{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cur.Close(context.TODO())

		var results []Request
		for cur.Next(context.TODO()) {
			var result Request
			err := cur.Decode(&result)
			if err != nil {
				log.Println(err)
				continue
			}
			results = append(results, result)
		}
		if err := cur.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(results)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	log.Println("listen :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}