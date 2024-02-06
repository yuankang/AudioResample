package main

import (
	"io/ioutil"
	"log"
	"os"
)

func SaveToDisk(name string, b []byte) (int, error) {
	f, err := os.Create(name)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer f.Close()

	n, err := f.Write(b)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return n, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	input, err := ioutil.ReadFile("audio48k.pcm")
	if err != nil {
		log.Fatalln(err)
	}

	res, err := New(48000, 8000, 1, I16, VeryHighQ)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Close()

	output, err := res.Write(input)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("len=%d, %x", len(output), output[:20])

	i, err := SaveToDisk("audio8kA.pcm", output)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("len=%d", i)

	output, err = res.Write(input)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("len=%d, %x", len(output), output[:20])

	i, err = SaveToDisk("audio8kB.pcm", output)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("len=%d", i)
}
