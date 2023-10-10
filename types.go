package main

type Response struct {
	Username      string `json:"username"`
	Authenticated bool   `json:"authenticated"`
}
