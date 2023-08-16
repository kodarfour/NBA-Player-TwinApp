package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/Kagami/go-face"
	"github.com/leandroveronezi/go-recognizer"
)

const player_headshots_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_headshots"
const models_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/models"
const player_data_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data"
const user_jpg_path = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/user image/user.jpg"

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

	username := "J. Cole"
	addFile(&rec, user_jpg_path, username)

	rec.SetSamples()

	playerName := "Tobias Harris"
	playerJPEG := playerName + ".jpeg"

	// INPUTTED User
	user_Descriptor, err := get_Descriptor(&rec, user_jpg_path)
	if err != nil {
		log.Fatal(err)
	}

	// SAME Player
	// user_Descriptor, err := get_Descriptor(&rec, filepath.Join(player_headshots_dir, playerJPEG))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// DIFFERENT Player
	// playerName2 := "Lebron James"
	// playerJPEG2 := playerName2 + ".jpeg"
	// user_Descriptor, err := get_Descriptor(&rec, filepath.Join(player_headshots_dir, playerJPEG2))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	player_descriptor, err := get_Descriptor(&rec, filepath.Join(player_headshots_dir, playerJPEG))
	if err != nil {
		log.Fatal(err)
	}
	euclideanD := face.SquaredEuclideanDistance(user_Descriptor, player_descriptor)
	rounded_euclideanD := fmt.Sprintf("%.2f", euclideanD)

	similarity_score := get_distance_based_similarity(euclideanD)
	similarity_score_percentage := fmt.Sprintf("%.2f", (math.Round(similarity_score*10000) / 100))

	fmt.Printf("Euclidean Distance between %s and %s is: %s\n", username, playerName, rounded_euclideanD)
	fmt.Printf("Similarity Percentage between %s and %s is: %s%%\n", username, playerName, similarity_score_percentage)

}

func get_Descriptor(rec *recognizer.Recognizer, jpg_Path string) (face.Descriptor, error) {

	thisFace, err := rec.Classify(jpg_Path)
	if err != nil {
		log.Fatal(err)
	}

	var this_Descriptor face.Descriptor
	for _, field := range thisFace {
		this_Descriptor = field.Descriptor
	}

	return this_Descriptor, err
}

func get_distance_based_similarity(euclidean_distance float64) float64 {
	denominator := float64(1) + euclidean_distance

	result := float64(1) / denominator

	return result
}

func addFile(rec *recognizer.Recognizer, Path, image_Id string) {

	err := rec.AddImageToDataset(Path, image_Id)

	if err != nil {
		fmt.Print("\n!!!!!!!!!!!!!!!! ERROR No face detected in image of:" + image_Id + " !!!!!!!!!!!!!!!!\n\n")
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
