package main

import "flag"

type config struct {
	port       int
	host       string
	Protocol   string
	rendezvous string
}

func parseFlags() *config {
	c := &config{}

	flag.StringVar(&c.host, "host", "0.0.0.0", "")
	flag.IntVar(&c.port, "port", 6666, "")
	flag.StringVar(&c.Protocol, "protocol", "/myChat/1.0", "")
	flag.StringVar(&c.rendezvous, "rendezvous", "meetme", "")
	flag.Parse()

	return c

}
