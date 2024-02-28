package domain

import "time"

type HTTPTransaction struct {
	Request  Request   `bson:"request"`
	Response Response  `bson:"response"`
	Time     time.Time `bson:"time"`
}

type Request struct {
	ID         string            `bson:"_id,omitempty"`
	Host       string            `bson:"host"`
	Method     string            `bson:"method"`
	Version    string            `bson:"version"`
	Path       string            `bson:"path"`
	Cookies    map[string]string `bson:"cookies, omitempty"`
	Headers    map[string]string `bson:"headers, omitempty"`
	GetParams  map[string]string `bson:"get_params, omitempty"`
	PostParams map[string]string `bson:"post_params, omitempty"`
}

type Response struct {
	StatusCode int               `bson:"status_code"`
	Headers    map[string]string `bson:"headers, omitempty"`
	Body       []byte            `bson:"body"`
}