package main

/* example code from https://golangbyexample.com/download-image-file-url-golang/ */
import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	fileName := "player_headshots/player.jpeg"
	URL := "https://cdn.nba.com/headshots/nba/latest/1040x760/1630534.png"
	err := downloadFile_at_headshots(URL, fileName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("File %s created", fileName)
}

func downloadFile_at_headshots(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
