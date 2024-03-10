package shema

type Connections struct {
	Login        string `json:"login"`
	Password     string `json:"password"`
	ConnectionDB string `json:"connection"`
}
