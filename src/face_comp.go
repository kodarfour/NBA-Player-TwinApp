package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/Kagami/go-face"
	"github.com/leandroveronezi/go-recognizer"
)

// A FloatHeap is a min-heap of float64s.
type FloatHeap []float64

func (h FloatHeap) Len() int           { return len(h) }
func (h FloatHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h FloatHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *FloatHeap) Push(x interface{}) {
	*h = append(*h, x.(float64))
}

func (h *FloatHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

const player_headshots_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_headshots"
const models_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/models"
const player_data_dir = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data"
const user_jpg_path = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/user image/user.jpg"
const populated_players_map_json_path = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data/playersmap.json"

func main() {

	rec := recognizer.Recognizer{}
	err := rec.Init(models_dir)

	if err != nil {
		fmt.Println(err)
		return
	}

	rec.Tolerance = 0.4
	rec.UseGray = true
	rec.UseCNN = true

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

	var players_euclideanD_map = make(map[string]float64)
	var players_map = make(map[string]interface{})

	path_to_player_dataset := filepath.Join(player_data_dir, "playerdataset.json")

	switch {
	case check_path(path_to_player_dataset):
		for _, player := range players {
			if player["nba-api-pID"] == nil {
				continue
			} else {
				player_name_str := fmt.Sprintf("%v", player["player-name"])
				players_map[player_name_str] = nil
			}
		}
		rec.LoadDataset(path_to_player_dataset)
		fmt.Println("Created players map and loaded playerdataset.json")

	default:
		for _, player := range players {
			if player["nba-api-pID"] == nil {
				continue
			} else {
				player_name_str := fmt.Sprintf("%v", player["player-name"])
				players_map[player_name_str] = nil
				player_name_jpeg := player_name_str + ".jpeg"
				add_file_to_dataset(&rec, filepath.Join(player_headshots_dir, player_name_jpeg), player_name_str)
			}
		}
		rec.SaveDataset(path_to_player_dataset)
		fmt.Println("Created players map and saved playerdataset.json")
	}

	rec.SetSamples()

	//username := "Kofi Darfour"
	user_Descriptor, err := get_Descriptor(&rec, user_jpg_path)
	if err != nil {
		log.Fatal(err)
	}

	if check_path(populated_players_map_json_path) {
		jsonFilePath := populated_players_map_json_path
		jsonFile, err := os.Open(jsonFilePath)
		if err != nil {
			log.Fatal("ERROR Couldnt open json: ", err)
		}
		defer jsonFile.Close()

		byteValueArray_for_jsonFile, _ := io.ReadAll(jsonFile)

		var players_map []map[string]interface{}

		json.Unmarshal(byteValueArray_for_jsonFile, &players_map)

		fmt.Println("Loaded playersmap.json and populated players map")
	} else {
		jsonFilePath := "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data/playerdataset.json"

		jsonFile, err := os.Open(jsonFilePath)
		if err != nil {
			log.Fatal("ERROR Couldnt open json: ", err)
		}

		defer jsonFile.Close()

		byteValueArray_for_jsonFile, _ := io.ReadAll(jsonFile)

		var dataset []map[string]interface{}

		json.Unmarshal(byteValueArray_for_jsonFile, &dataset)

		for _ , dict := range dataset {
			str := convert_to_string(dict["Id"])
			descriptor := dict["Descriptor"]
			fmt.Println(descriptor)
			players_map[str] = descriptor
		}

		file, err := json.MarshalIndent(players_map, "", "\t")
		if err != nil {
			log.Fatal("ERROR:", err)
		}

		os.WriteFile(populated_players_map_json_path, file, 0644)

		fmt.Println("Saved playersmap.json and populated players map")
	}

	for player, descriptor := range players_map {
		player_descriptor := convert_to_descriptor(descriptor)
		euclideanD := face.SquaredEuclideanDistance(user_Descriptor, player_descriptor)
		players_euclideanD_map[player] = euclideanD
	}

	fmt.Println("Players Euclidean Distance Mapped")

	get_top_n_closest_players(5, players_euclideanD_map)

	// playerName := "Bruno Fernando"
	// playerJPEG := playerName + ".jpeg"

	// player_descriptor, err := get_Descriptor(&rec, filepath.Join(player_headshots_dir, playerJPEG))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// euclideanD := face.SquaredEuclideanDistance(user_Descriptor, player_descriptor)

	// rounded_euclideanD := fmt.Sprintf("%.5f", euclideanD)
	// similarity_score := get_distance_based_similarity(euclideanD)
	// similarity_score_percentage := fmt.Sprintf("%.2f", (math.Round(similarity_score*10000) / 100))

	// fmt.Printf("Euclidean Distance between %s and %s is: %s\n", username, playerName, rounded_euclideanD)
	// fmt.Printf("Similarity Percentage between %s and %s is: %s%%\n", username, playerName, similarity_score_percentage)

}

func get_top_n_closest_players(n int, players_ed_map map[string]float64) {
	top_n_heap := &FloatHeap{}
	heap.Init(top_n_heap)

	for player, euclidean_distance := range players_ed_map {
		fmt.Println(player, euclidean_distance)
		if top_n_heap.Len() < 5 {
			heap.Push(top_n_heap, euclidean_distance)
			fmt.Println(top_n_heap)
		} else if (*top_n_heap)[0] < euclidean_distance {
			(*top_n_heap)[0] = euclidean_distance
		}
	}
	fmt.Println(top_n_heap)

}

func convert_to_float64(object interface{}) float64 {
	euclidean_distance, _ := object.(float64)
	return euclidean_distance
}


func convert_to_string(object interface{}) string {
	name, _ := object.(string)
	return name
}

func convert_to_descriptor(object interface{}) face.Descriptor {
	descriptor, _ := object.(face.Descriptor)
	return descriptor
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

func add_file_to_dataset(rec *recognizer.Recognizer, Path, image_Id string) {

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
