package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	earthRadius    = 6371.0
	baseURL        = "https://nominatim.openstreetmap.org/search"
	requestTimeout = 5 * time.Second
	maxRetries     = 3
)

type Location struct {
	Latitude  float64
	Longitude float64
}

type ResponseLocation []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type Request struct {
	From  string  `json:"from"`
	To    string  `json:"to"`
	Range float64 `json:"range"`
}

type Response struct {
	KM      string `json:"km"`
	InRange bool   `json:"inRange"`
}

type CEP string

func NewCEP(cep string) (CEP, error) {
	if len(cep) != 8 {
		return "", fmt.Errorf("CEP deve conter 8 dígitos")
	}
	if !regexp.MustCompile(`^\d{8}$`).MatchString(cep) {
		return "", fmt.Errorf("CEP deve conter apenas números")
	}
	return CEP(cep), nil
}

func getLatitudeLongitude(cep CEP) (float64, float64, error) {
	params := url.Values{}
	params.Set("postalcode", string(cep))
	params.Set("country", "Brazil")
	params.Set("format", "json")

	client := &http.Client{Timeout: requestTimeout}

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := client.Get(baseURL + "?" + params.Encode())
		if err != nil {
			if attempt == maxRetries-1 {
				return 0, 0, fmt.Errorf("falha ao consultar API após %d tentativas: %v", maxRetries, err)
			}
			time.Sleep(time.Second * time.Duration(attempt+1))
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, 0, fmt.Errorf("erro ao ler resposta: %v", err)
		}

		var result ResponseLocation
		if err := json.Unmarshal(body, &result); err != nil {
			return 0, 0, fmt.Errorf("erro ao decodificar resposta: %v", err)
		}

		if len(result) == 0 {
			return 0, 0, fmt.Errorf("não foi possível encontrar coordenadas para o CEP %s", cep)
		}

		lat, err := strconv.ParseFloat(result[0].Lat, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("erro ao converter latitude: %v", err)
		}

		lon, err := strconv.ParseFloat(result[0].Lon, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("erro ao converter longitude: %v", err)
		}

		return lat, lon, nil
	}

	return 0, 0, fmt.Errorf("número máximo de tentativas excedido")
}

func haversine(fromLocation, toLocation Location, maxDistance float64) (float64, bool) {
	lat1Rad := fromLocation.Latitude * math.Pi / 180
	lon1Rad := fromLocation.Longitude * math.Pi / 180
	lat2Rad := toLocation.Latitude * math.Pi / 180
	lon2Rad := toLocation.Longitude * math.Pi / 180

	dlon := lon2Rad - lon1Rad
	dlat := lat2Rad - lat1Rad

	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c

	return distance, distance <= maxDistance
}

func getCoordinates(from string, to string, maxRange float64) (float64, bool, error) {
	log.Printf("Iniciando busca de coordenadas para CEPs: %s -> %s", from, to)

	fromCEP, err := NewCEP(from)
	if err != nil {
		return 0, false, fmt.Errorf("CEP de origem inválido: %v", err)
	}

	toCEP, err := NewCEP(to)
	if err != nil {
		return 0, false, fmt.Errorf("CEP de destino inválido: %v", err)
	}

	fromLocation := Location{}
	fromLocation.Latitude, fromLocation.Longitude, err = getLatitudeLongitude(fromCEP)
	if err != nil {
		return 0, false, fmt.Errorf("erro ao buscar coordenadas do CEP de origem: %v", err)
	}

	toLocation := Location{}
	toLocation.Latitude, toLocation.Longitude, err = getLatitudeLongitude(toCEP)
	if err != nil {
		return 0, false, fmt.Errorf("erro ao buscar coordenadas do CEP de destino: %v", err)
	}

	distance, inRange := haversine(fromLocation, toLocation, maxRange)
	return distance, inRange, nil
}

func main() {
	f, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Erro ao abrir arquivo de log: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	router := gin.Default()

	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	router.POST("/cep", func(c *gin.Context) {
		var request Request
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Printf("Erro ao processar JSON da requisição: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		distance, inRange, err := getCoordinates(request.From, request.To, request.Range)
		if err != nil {
			log.Printf("Erro ao processar coordenadas: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := Response{
			KM:      fmt.Sprintf("%.2f", distance),
			InRange: inRange,
		}
		c.JSON(http.StatusOK, response)
	})

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
