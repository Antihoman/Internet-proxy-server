package domain

import "time"

type HTTPTransaction struct {
	ID       interface{} `bson:"_id,omitempty"`
	Request  Request     `bson:"request"`
	Response Response    `bson:"response"`
	Time     time.Time   `bson:"time"`
}

type Request struct {
	Host       string            `bson:"host"`
	Method     string            `bson:"method"`
	Version    string            `bson:"version"`
	Path       string            `bson:"path"`
	Protocol   string            `bson:"protocol"`
	Cookies    map[string]string `bson:"cookies, omitempty"`
	Headers    map[string]string `bson:"headers, omitempty"`
	GetParams  map[string]string `bson:"get_params, omitempty"`
	PostParams map[string]string `bson:"post_params, omitempty"`
	RawBody    []byte            `bson:"raw_body"`
}

type Response struct {
	StatusCode    int               `bson:"status_code"`
	Headers       map[string]string `bson:"headers, omitempty"`
	ContentLenght int               `bson:"content_length"`
	RawBody       []byte            `bson:"raw_body"`
	TextBody      string            `bson:"text_body"`
}