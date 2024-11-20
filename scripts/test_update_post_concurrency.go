package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type UpdatedPostPayload struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

func updatePost(postId int64, p UpdatedPostPayload, wg *sync.WaitGroup) {
	url := fmt.Sprintf("http://localhost:3002/v1/posts/%d", postId)
	fmt.Println(url)
	b, _ := json.Marshal(p)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(b))
	if err != nil {
		log.Printf("error creating request : %+v", err)
		wg.Done()
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resq, err := client.Do(req)
	if err != nil {
		log.Printf("error sending request : %+v", err)
		wg.Done()
		return
	}
	defer resq.Body.Close()
	fmt.Printf("updated response status : %+v", resq.Status)
	wg.Done()

}

func main() {
	var wg sync.WaitGroup
	postId := 3
	wg.Add(1)
	content := "New Content from User B "
	// title := "New title from User A "

	// go updatePost(int64(postId), UpdatedPostPayload{Title: title}, &wg)
	go updatePost(int64(postId), UpdatedPostPayload{Content: content}, &wg)

	wg.Wait()
}
