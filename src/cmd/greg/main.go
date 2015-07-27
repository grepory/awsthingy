package main

import (
	"log"
	"os"

	"client"
)

func main() {
	url := os.Getenv("URL")
	vpc := os.Getenv("VPC")
	aws_access_key_id := os.Getenv("AWS_ACCESS_KEY_ID")
	aws_secret_key := os.Getenv("AWS_SECRET_ACCESS_KEY")

	remote, err := client.NewClient(url, aws_access_key_id, aws_secret_key)

	instances, err := remote.ListInstances(vpc)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Print(instances)
}
