package main

type Config struct {
	Src ConfigAndGroup
	Dst []ConfigAndGroup
}

type ConfigAndGroup struct {
	Config string
	Group  string
}
