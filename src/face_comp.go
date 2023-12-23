package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"
	"strings"

	"github.com/Kagami/go-face"
	"github.com/leandroveronezi/go-recognizer"
)

const player_headshots_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_headshots"
const models_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/models"
const player_data_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data"
const user_jpg_path = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/user image"
const populated_players_map_json_path = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data/playersmap.json"

func convert_to_float32(object interface{}) ([128]float32, error) {
	slice, ok := object.([]interface{})
	if !ok || len(slice) != 128 {
		return [128]float32{}, fmt.Errorf("ERROR: Data is not a slice of interfaces or has incorrect length")
	}

	var result [128]float32

	for i, v := range slice {
		value, ok := v.(float64)
		if !ok {
			return [128]float32{}, fmt.Errorf("ERROR: Element at index %d is not a float64", i)
		}
		result[i] = float32(value)
	}

	return result, nil
}

func convert_to_string(object interface{}) string {
	name, _ := object.(string)
	return name
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

func get_averageDescriptor(descriptors []face.Descriptor) face.Descriptor {
	if len(descriptors) == 0 {
		return face.Descriptor{}
	}

	var averageDescriptor face.Descriptor
	for _, descriptor := range descriptors {
		for i, value := range descriptor {
			averageDescriptor[i] += value
		}
	}

	for i := range averageDescriptor {
		averageDescriptor[i] /= float32(len(descriptors))
	}

	return averageDescriptor
}

func get_distance_based_similarity(euclidean_distance float64) float64 {
	denominator := float64(1) + euclidean_distance

	result := float64(1) / denominator

	return result
}

func add_file_to_dataset(rec *recognizer.Recognizer, Path, image_Id string) {

	err := rec.AddImageToDataset(Path, image_Id)

	if err != nil {
		fmt.Print("\nERROR: No face detected in image of \"" + getLastDir(Path) + "\"\n\n")
		return
	} else {
		fmt.Println("Added \"" + getLastDir(Path) + "\" to data set")
	}
}

func check_path(pathname string) bool {
	info, err := os.Stat(pathname)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func check_image(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.HasSuffix(lower, ".jpeg") // only works wit jpg/jpeg files
}

func getLastDir(path string) string {
	return filepath.Base(path)
}

func main() {
	rec := recognizer.Recognizer{}
	err := rec.Init(models_dir)

	if err != nil {
		fmt.Println(err)
		return
	}

	rec.Tolerance = 0.3
	rec.UseGray = false
	rec.UseCNN = true

	defer rec.Close()

	fmt.Println("go-recoginzer succesfully initialized...")

	jsonFilePath := "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data/playerdata.json"

	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		log.Fatal("ERROR: (Couldnt open json) ", err)
	}

	fmt.Println("Successfully opened the playerdata json file...")

	defer jsonFile.Close()

	byteValueArray_for_jsonFile, _ := io.ReadAll(jsonFile)

	var players []map[string]interface{}

	json.Unmarshal(byteValueArray_for_jsonFile, &players)

	var players_map = make(map[string][]face.Descriptor)

	path_to_player_dataset := filepath.Join(player_data_dir, "playerdataset.json")

	switch {
	case check_path(path_to_player_dataset):
		rec.LoadDataset(path_to_player_dataset)
		fmt.Println("Loaded playerdataset.json...")

	default:
		// loop through folders in /player_headshots
		err := filepath.Walk(player_headshots_dir, func(currentPlayer_dir string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// checks if dir exists
			if info.IsDir() && currentPlayer_dir != player_headshots_dir {
				// loops through files
				filepath.Walk(currentPlayer_dir, func(currentPlayer_image string, innerInfo os.FileInfo, innerErr error) error {
					if innerErr != nil {
						return innerErr
					}

					// Check if the file is an image ending in .jpeg
					if !innerInfo.IsDir() && check_image(innerInfo.Name()) {
						add_file_to_dataset(&rec, currentPlayer_image, getLastDir(currentPlayer_dir))
					}

					return nil
				})
			}

			return nil
		})

		if err != nil {
			fmt.Println("ERROR: (walking the path)", err)
		}

	}
	rec.SaveDataset(path_to_player_dataset)
	fmt.Println("Saved playerdataset.json...")

	rec.SetSamples()

	if check_path(populated_players_map_json_path) {
		jsonFilePath := populated_players_map_json_path
		jsonFile, err := os.Open(jsonFilePath)
		if err != nil {
			log.Fatal("ERROR: (Couldnt open json) ", err)
		}
		defer jsonFile.Close()

		byteValueArray_for_jsonFile, _ := io.ReadAll(jsonFile)

		json.Unmarshal(byteValueArray_for_jsonFile, &players_map)

		fmt.Println("Loaded playersmap.json...")
	} else {
		jsonFilePath := "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data/playerdataset.json"

		jsonFile, err := os.Open(jsonFilePath)
		if err != nil {
			log.Fatal("ERROR: (Couldnt open json) ", err)
		}

		defer jsonFile.Close()

		byteValueArray_for_jsonFile, _ := io.ReadAll(jsonFile)

		var dataset []map[string]interface{}

		json.Unmarshal(byteValueArray_for_jsonFile, &dataset)

		for _, dict := range dataset {
			str := convert_to_string(dict["Id"])
			descriptor, err := convert_to_float32(dict["Descriptor"])
			if err != nil {
				fmt.Println("ERROR: (converting to [128]float3)", err)
				return
			}
			players_map[str] = append(players_map[str], (face.Descriptor)(descriptor))
			fmt.Println("\n", str)
			fmt.Println("\n", players_map[str])
		}

		file, err := json.MarshalIndent(players_map, "", "\t")
		if err != nil {
			log.Fatal("ERROR:", err)
		}

		os.WriteFile(populated_players_map_json_path, file, 0644)

		fmt.Println("Saved playersmap.json...")
	}

	playerName := "Joel Embiid" //random_player(players_map)

	username := "Danielle"

	user_Descriptor, err := get_Descriptor(&rec, filepath.Join(user_jpg_path, "danielle.jpg"))
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("Succesfully retrieved %s's facial descriptors...\n", username)
	}
	
	defer rec.Close()

	euclideanD := face.SquaredEuclideanDistance(user_Descriptor, get_averageDescriptor(players_map[playerName]))

	rounded_euclideanD := fmt.Sprintf("%.5f", euclideanD)
	similarity_score := get_distance_based_similarity(euclideanD)
	similarity_score_percentage := fmt.Sprintf("%.2f", (math.Round(similarity_score*10000) / 100))

	fmt.Printf("Euclidean Distance between %s and %s is: %s\n", username, playerName, rounded_euclideanD)
	fmt.Printf("Similarity Percentage between %s and %s is: %s%%\n", username, playerName, similarity_score_percentage)
}

func random_player(m map[string][]face.Descriptor) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys[r.Intn(len(keys))]
}