package domain

import "time"

type HTTPTransaction struct {
	ID       interface{} `bson:"_id,omitempty" json:"_id,omitempty"`
	Request  Request     `bson:"request" json:"request"`
	Response Response    `bson:"response" json:"response"`
	Time     time.Time   `bson:"time" json:"time"`
}

type Request struct {
	Host       string            `bson:"host" json:"host"`
	Method     string            `bson:"method" json:"method"`
	Version    string            `bson:"version" json:"version"`
	Path       string            `bson:"path" json:"path"`
	Protocol   string            `bson:"protocol" json:"protocol"`
	Cookies    map[string]string `bson:"cookies" json:"cookies"`
	Headers    map[string]string `bson:"headers" json:"headers"`
	GetParams  map[string]string `bson:"get_params" json:"get_params"`
	PostParams map[string]string `bson:"post_params" json:"post_params"`
	RawBody    []byte            `bson:"raw_body" json:"raw_body"`
}

type Response struct {
	StatusCode    int               `bson:"status_code" json:"status_code"`
	Headers       map[string]string `bson:"headers" json:"headers"`
	ContentLenght int               `bson:"content_length" json:"content_length"`
	RawBody       []byte            `bson:"raw_body" json:"raw_body"`
	TextBody      string            `bson:"text_body" json:"text_body"`
}