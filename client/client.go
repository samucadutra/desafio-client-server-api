package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type PriceResponse struct {
	Bid string `json:"bid"`
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Error:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var priceResponse PriceResponse
	if err := json.Unmarshal(body, &priceResponse); err != nil {
		panic(err)
	}
	println(string(body))
	salvarArquivoCotacao(priceResponse)
}

func salvarArquivoCotacao(priceResponse PriceResponse) {
	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}

	tamanho, err := f.WriteString("Dólar: " + priceResponse.Bid)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Arquivo criado com sucesso! Tamanho: %d bytes\n", tamanho)
	f.Close()
}
