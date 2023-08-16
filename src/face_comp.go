package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/leandroveronezi/go-recognizer"
)

const player_headshots_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_headshots"
const models_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/models"
const player_data_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data"

func main() {
	rec := recognizer.Recognizer{}
	err := rec.Init(models_dir)

	if err != nil {
		fmt.Println(err)
		return
	}

	rec.Tolerance = 0.4
	rec.UseGray = true
	rec.UseCNN = false

	defer rec.Close()

	fmt.Println("go-recoginzer succesfully initialized...")

	jsonFilePath := "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data/playerdata.json"

	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		log.Fatal("ERROR Couldnt open json: ", err)
	}

	fmt.Println("Successfully opened the playerdata json file...")

	defer jsonFile.Close()

	byteValueArray_for_jsonFile, _ := io.ReadAll(jsonFile)

	var players []map[string]interface{}

	json.Unmarshal(byteValueArray_for_jsonFile, &players)

	var players_slice []string

	path_to_player_dataset := filepath.Join(player_data_dir, "playerdataset.json")

	switch {
	case check_path(path_to_player_dataset):
		for _, player := range players {
			if player["nba-api-pID"] == nil {
				continue
			} else {
				player_name_str := fmt.Sprintf("%v", player["player-name"])
				players_slice = append(players_slice, player_name_str)
			}
		}
		rec.LoadDataset(path_to_player_dataset)
		fmt.Println("Populated players slice and loaded playerdataset.json")

	default:
		for _, player := range players {
			if player["nba-api-pID"] == nil {
				continue
			} else {
				player_name_str := fmt.Sprintf("%v", player["player-name"])
				players_slice = append(players_slice, player_name_str)
				player_name_jpeg := player_name_str + ".jpeg"
				addFile(&rec, filepath.Join(player_headshots_dir, player_name_jpeg), player_name_str)
			}
		}
		rec.SaveDataset(path_to_player_dataset)
		fmt.Println("Populated players slice and saved playerdataset.json")
	}

	rec.SetSamples()

}

func addFile(rec *recognizer.Recognizer, Path, image_Id string) {

	err := rec.AddImageToDataset(Path, image_Id)

	if err != nil {
		fmt.Print("\n!!!!!!!!!!!!!!!! ERROR No face detected in: " + image_Id + ".jpeg !!!!!!!!!!!!!!!!\n\n")
		return
	} else {
		fmt.Println("Added " + image_Id + " to data set")
	}
}

func check_path(pathname string) bool {
	info, err := os.Stat(pathname)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
