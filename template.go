package main

type Template struct {
	Bastion  *Bastion `json:"bastion"`
	Remote   *Remote  `json:"remote"`
	Commands []string `json:"commands"`
}

type Bastion struct {
	Host string `json:"host"`
	User string `json:"user"`
}

type Aws struct {
	Nameprefix string `json:"name-prefix"`
}

type Remote struct {
	Hosts []string `json:"hosts"`
	Aws   *Aws     `json:"aws"`
	User  string   `json:"user"`
}
