package domain

import "time"

type Request struct {
	ID      string            `bson:"_id,omitempty"`
	Host    string            `bson:"host"`
	Method  string            `bson:"method"`
	Version string            `bson:"version"`
	Path    string            `bson:"path"`
	Headers map[string]string `bson:"headers, omitempty"`
	Time    time.Time         `bson:"time"`
}