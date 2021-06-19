package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type TablaDataset struct {
	IdPersona  string `json:"IdPersona"`
	Prestacion string `json:"Prestacion"`
	TipoOtra   string `json:"TipoOtra"`
	TipoBiente string `json:"TipoBien"`
	Beneficio  string `json:"Beneficio"`
}
type Kmeans struct {
	data                [][]float64
	data_labels         []int
	data_representativa [][]float64
}

var Tablas []TablaDataset

func Transp(source [][]float64) [][]float64 {
	var dest [][]float64
	for i := 0; i < len(source[0]); i++ {
		var temp []float64
		for j := 0; j < len(source); j++ {
			temp = append(temp, 0.0)
		}
		dest = append(dest, temp)
	}

	for i := 0; i < len(source); i++ {
		for j := 0; j < len(source[0]); j++ {
			dest[j][i] = source[i][j]
		}
	}
	return dest
}

func Dist(source, dest []float64) float64 {
	var dist float64
	for i := range source {
		dist += math.Pow(source[i]-dest[i], 2)
	}
	return math.Sqrt(dist)
}

func Minimo(target []float64) int {
	var (
		index int
		base  float64
	)
	for i, d := range target {
		if i == 0 {
			index = i
			base = d
		} else {
			if d < base {
				index = i
				base = d
			}
		}

	}
	return index
}

func (km *Kmeans) fit(X [][]float64, k int) []int {
	rand.Seed(time.Now().UnixNano())
	km.data = X

	transposedData := Transp(km.data)
	var minMax [][]float64
	for _, d := range transposedData {
		var (
			min float64
			max float64
		)
		for _, n := range d {
			min = math.Min(min, n)
			max = math.Max(max, n)
		}
		minMax = append(minMax, []float64{min, max})
	}
	for i := 0; i < k; i++ {
		km.data_representativa = append(km.data_representativa, make([]float64, len(minMax)))
	}
	for i := 0; i < k; i++ {
		for j := 0; j < len(minMax); j++ {
			km.data_representativa[i][j] = rand.Float64()*(minMax[j][1]-minMax[j][0]) + minMax[j][0]
		}
	}
	for _, d := range km.data {
		var distance []float64
		for _, r := range km.data_representativa {
			distance = append(distance, Dist(d, r))
		}
		km.data_labels = append(km.data_labels, Minimo(distance))
	}
	for {
		var tempRepresentatives [][]float64
		for i := range km.data_representativa {
			var grouped [][]float64
			for j, d := range km.data {
				if km.data_labels[j] == i {
					grouped = append(grouped, d)
				}
			}
			if len(grouped) != 0 {

				transposedGroup := Transp(grouped)
				updated := []float64{}
				for _, vectors := range transposedGroup {

					value := 0.0
					for _, v := range vectors {
						value += v
					}
					updated = append(updated, value/float64(len(vectors)))
				}
				tempRepresentatives = append(tempRepresentatives, updated)
			}
		}
		km.data_representativa = tempRepresentatives

		tempLabel := []int{}
		for _, d := range km.data {
			var distance []float64
			for _, r := range km.data_representativa {
				distance = append(distance, Dist(d, r))
			}
			tempLabel = append(tempLabel, Minimo(distance))
		}
		if reflect.DeepEqual(km.data_labels, tempLabel) {
			break
		} else {
			km.data_labels = tempLabel
		}
	}
	return km.data_labels
}

func MakeAlgorithm(w http.ResponseWriter, r *http.Request) {
	url := "https://raw.githubusercontent.com/JorgeDanielVital/data/master/data.csv"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	reader.Comma = ','
	datas, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
	}

	X := [][]float64{}
	Y := []string{}
	for _, info := range datas {
		temp := []float64{}
		for _, i := range info[:4] {
			parsedValue, err := strconv.ParseFloat(i, 64)
			if err != nil {
				panic(err)
			}
			temp = append(temp, parsedValue)
		}
		X = append(X, temp)
		Y = append(Y, info[4])
	}
	km := Kmeans{}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(km.fit(X, 4))
	json.NewEncoder(w).Encode(Y)
}

func getTable(w http.ResponseWriter, r *http.Request) {
	url := "https://raw.githubusercontent.com/JorgeDanielVital/data/master/data.csv"
	resp, _ := http.Get(url)
	reader := csv.NewReader(resp.Body)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	reader.Comma = ';'
	datas, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	for _, line := range datas {
		IdPersona := strings.Split(line[0], ",")[0]
		Prestacion := strings.Split(line[0], ",")[1]
		TipoOtra := strings.Split(line[0], ",")[2]
		TipoBiente := strings.Split(line[0], ",")[3]
		Beneficio := strings.Split(line[0], ",")[4]
		tabla := TablaDataset{IdPersona: IdPersona, Prestacion: Prestacion, TipoOtra: TipoOtra, TipoBiente: TipoBiente, Beneficio: Beneficio}
		Tablas = append(Tablas, tabla)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Tablas)
}

func addData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var tableData TablaDataset
	_ = json.NewDecoder(r.Body).Decode(&tableData)
	Tablas = append(Tablas, tableData)
	MakeAlgorithm(w, r)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", getTable).Methods("GET")
	r.HandleFunc("/KNN", MakeAlgorithm).Methods("GET")
	r.HandleFunc("/Add", addData).Methods("POST")
	log.Fatal(http.ListenAndServe(":4000", r))
}
