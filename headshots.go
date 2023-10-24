package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/caffix/cloudflare-roundtripper/cfrt"
	"github.com/gocolly/colly/v2"
)

func main() {

	jsonFilePath := "player_data/playerdata.json"

	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		log.Fatal("ERROR Couldnt open json: ", err)
	}

	fmt.Println("Successfully opened the json file...")

	defer jsonFile.Close()

	byteValueArray_for_jsonFile, _ := io.ReadAll(jsonFile)

	var players []map[string]interface{}

	json.Unmarshal(byteValueArray_for_jsonFile, &players)

	client := createClient()

	for _, player := range players {
		if player["nba-api-pID"] == nil {
			continue
		} else {
			player_name_str := fmt.Sprintf("%v", player["player-name"])
			c := colly.NewCollector()
			player_name_str_modified1 := strings.ReplaceAll(player_name_str, " ", "-")
			player_name_str_modified2 := strings.ReplaceAll(player_name_str, " ", "%20") + "%20"
			gettyURL := fmt.Sprintf("https://www.gettyimages.com/photos/%s-basketball?page=1&assettype=image&compositions=lookingatcamera&family=editorial&numberofpeople=one&phrase=%sBasketball&sort=newest", player_name_str_modified1, player_name_str_modified2)
			raw_player_ID := player["nba-api-pID"]
			player_id_float, ok := raw_player_ID.(float64) // due to weird interface conversion stuff needs to be made a float
			if ok {
				var player_ID_int int = int(player_id_float) // then converted to int
				player_id_str := strconv.Itoa(player_ID_int) 
				player_headshot_path := "player_headshots/" + player_name_str
				if check_path(player_headshot_path) {
					fmt.Printf("Folder for %s already exists\n", player_name_str)
					continue
				} else {
					directory := filepath.Join("/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_headshots", player_name_str)
					err = os.MkdirAll(directory, 0o755)

					if err != nil {
						fmt.Printf("Failed to create directory for %s: %v\n", player_name_str, err)
					} else {
						fmt.Printf("\n\nFOLDER CREATED FOR %s\n", player_name_str)
					}

					var count int = 0

					c.OnHTML("img", func(e *colly.HTMLElement) {
						if count >= 50 { // want up to 50 images from Getty
							return
						}
						fmt.Print(gettyURL)
						link := e.Attr("src")
						imgPath := fmt.Sprintf(player_headshot_path+"/%s%d.jpeg", player_name_str, count)

						err := downloadGettyImage(client, link, imgPath)
						if err != nil {
							log.Println("Error downloading image for:", player_name_str, err)
						} else {
							fmt.Printf("\n\nGETTY IMAGE #%d CREATED FOR %s\n\n", count, player_name_str)
						}
						count++
					})

					c.OnError(func(r *colly.Response, err error) {
						if r.StatusCode == 404 {
							return
						}
						log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
					})

					err = c.Visit(gettyURL)
					if err != nil {
						log.Printf("Failed to visit Getty Images for %s: %s", player_name_str, err)
					} else {
						fmt.Print("\n\n",gettyURL,"\n\n")
					}

					jordan_clarkson_URL := "https://images.seattletimes.com/wp-content/uploads/2022/09/urn-publicid-ap-org-4a66b28596900a5159e6dde6d294d216Jazz_Media_Day_Basketball_04107.jpg?d=1020x680"
					jarace_walker_URL := "https://a.espncdn.com/combiner/i?img=/i/headshots/mens-college-basketball/players/full/5106060.png&w=350&h=254"
					playerURL := "https://cdn.nba.com/headshots/nba/latest/1040x760/" + player_id_str + ".png"

					switch {
					case player_name_str == "Jordan Clarkson":
						// go-recognizer having issues with the nba.com source
						err := downloadFile_at_headshots(jordan_clarkson_URL, player_headshot_path+"/"+player_name_str+".jpeg")
						if err != nil {
							log.Println("Jordan Clarkson FAILED", err)
						}
					case player_name_str == "Jarace Walker":
						// not really necesarry as he now has a player image on nba.com but i'll keep it anyways
						err := downloadFile_at_headshots(jarace_walker_URL, player_headshot_path+"/"+player_name_str+".jpeg")
						if err != nil {
							log.Println("Jarace Walker FAILED", err)
						}
					default:
						err := downloadFile_at_headshots(playerURL, player_headshot_path+"/"+player_name_str+".jpeg")
						if err != nil {
							log.Println(err)
						} else {
							fmt.Printf("COMPLETED ALL JPEG for %s\n", player_name_str)
						}
					}
				}
			} else {
				log.Fatal(ok)
			}
		}
	}
}

func createClient() *http.Client { // in order to avoid Error: 400s/ratelimits
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

	client.Transport, _ = cfrt.New(client.Transport)
	return client
}

func downloadGettyImage(client *http.Client, url, filename string) error {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", 
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func downloadFile_at_headshots(URL, fileName string) error {

	var err error

	client := &http.Client{ // no need to reuse in this case
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
	req.Header.Set("User-Agent", 
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537")

	response, err := client.Do(req)

	if response.StatusCode != 200 {
		log.Fatal(response.StatusCode)
		//errors.New("Received non 200 response code")
	}

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
	return info.IsDir()
}
