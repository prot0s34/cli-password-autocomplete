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

type FolderResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Object string `json:"object"`
		Data   []struct {
			Object string `json:"object"`
			ID     string `json:"id"`
			Name   string `json:"name"`
		} `json:"data"`
	} `json:"data"`
}

type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
