package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/caffix/cloudflare-roundtripper/cfrt"
)

func main() {
	jsonFilePath := "player_data/playerdata.json"

	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		log.Fatal("Couldnt open json: ", err)
	}

	fmt.Println("Opened the json file")

	defer jsonFile.Close()

	byteValueArray_for_jsonFile, _ := io.ReadAll(jsonFile)

	var players []map[string]interface{}

	json.Unmarshal(byteValueArray_for_jsonFile, &players)

	var players_slice []string

	for _, player := range players {
		if player["nba-api-pID"] == nil {
			continue
		} else {
			player_name_str := fmt.Sprintf("%v", player["player-name"])
			player_name_folder_path := "player_headshots/" + player_name_str
			if check_path(player_name_folder_path) {
				continue
			} else {
				err := os.MkdirAll(player_name_folder_path, 0777)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			raw_player_ID := player["nba-api-pID"]
			player_id_float, ok := raw_player_ID.(float64)
			if ok {
				var player_ID_int int = int(player_id_float)
				player_id_str := strconv.Itoa(player_ID_int)
				player_headshot_path := player_name_folder_path + "/" + player_name_str + ".jpeg"

				if check_path(player_headshot_path) {
					fmt.Printf("JPEG for %s already exists\n", player_name_str)
					players_slice = append(players_slice, player_name_str)
					continue
				} else {
					players_slice = append(players_slice, player_name_str)
					URL := "https://cdn.nba.com/headshots/nba/latest/1040x760/" + player_id_str + ".png"
					err := downloadFile_at_headshots(URL, player_headshot_path)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Printf("JPEG for %s created\n", player_name_str)
				}
			} else {
				log.Fatal(ok)
			}
		}
	}

	fmt.Println(players_slice)
}

func downloadFile_at_headshots(URL, fileName string) error {
	var err error

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
				DualStack: true,
			}).DialContext,
		},
	}

	client.Transport, err = cfrt.New(client.Transport)
	if err != nil {
		return err
	}

	//Get the response bytes from the url
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return err
	}

	response, err := client.Do(req)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Fatal(response.StatusCode)
		//errors.New("Received non 200 response code")
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

func check_path(pathname string) bool {
	info, err := os.Stat(pathname)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
