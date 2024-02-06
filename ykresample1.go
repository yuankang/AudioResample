package main

import (
	"io/ioutil"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	input, err := ioutil.ReadFile("xxx.pcm")
	if err != nil {
		log.Fatalln(err)
	}

	res, err := New(48000, 8000, 1, 3, 6)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Close()

	i, err, output := res.Write(input)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("len, %x", len(output), output[:20])
}
