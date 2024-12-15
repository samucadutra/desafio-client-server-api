package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Cotacao struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/", BuscaCotacaoDolarHandler)
	http.ListenAndServe(":8080", nil)
}

func BuscaCotacaoDolarHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	cotacao, err := buscaCotacaoDolarApiExterna(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = salvarCotacaoBancoDeDados(ctx, cotacao)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(cotacao)

	//select {
	//case <-ctx.Done():
	//	fmt.Println("Get  cancelled. Timeout reached.")
	//	return
	//case <-time.After(1 * time.Second):
	//	fmt.Println("Hotel booked.")
	//}
}

func buscaCotacaoDolarApiExterna(ctx context.Context) (*Cotacao, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		return nil, err
	}
	return &cotacao, nil
}

func salvarCotacaoBancoDeDados(ctx context.Context, cotacao *Cotacao) error {
	fmt.Println("Salvando a cotação do ativo: ", cotacao.Usdbrl.Name, " no valor de: ", cotacao.Usdbrl.Bid)
	return nil
}
