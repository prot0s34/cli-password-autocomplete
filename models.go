package main

type Item struct {
	Name  string `json:"name"`
	Login struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"login"`
}

type ListResponse struct {
	Data struct {
		Items []Item `json:"data"`
	} `json:"data"`
}
