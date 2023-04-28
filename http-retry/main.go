package main

import (
	"fmt"
	"io/ioutil"

	rhttp "github.com/kdkumawat/golang/http-retry/http"
)

func main() {
	client := rhttp.NewRetryableClient()
	resp, err := client.Get("https://reqres.in/api/users/2")
	if err != nil {
		fmt.Println("err", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("err reading body : %w", err)
	}

	fmt.Println("resp", string(body))
}
