package main

import (
	"context"
	"encoding/json"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
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

type cotacaoEntity struct {
	gorm.Model
	Code       string
	Codein     string
	Name       string
	High       string
	Low        string
	VarBid     string
	PctChange  string
	Bid        string
	Ask        string
	Timestamp  string
	CreateDate string
}

type PriceResponse struct {
	Bid string `json:"bid"`
}

func main() {
	db := initDB()
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		cotacao, err := buscaCotacaoDolarApiExterna()
		if err != nil {
			if err.Error() == context.DeadlineExceeded.Error() {
				log.Println("Erro de timeout ao buscar cotação: ", err)
				w.WriteHeader(http.StatusRequestTimeout)
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err = salvarCotacaoBancoDeDados(db, cotacao)
		if err != nil {
			if err.Error() == context.DeadlineExceeded.Error() {
				log.Println("Erro de timeout ao salvar cotação no banco de dados: ", err)
				w.WriteHeader(http.StatusRequestTimeout)
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(PriceResponse{Bid: cotacao.Usdbrl.Bid})
	})

	http.ListenAndServe(":8080", mux)

}

func initDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}
	err = db.AutoMigrate(&cotacaoEntity{})
	if err != nil {
		log.Fatal("Error migrating DB: ", err)
	}
	return db
}

func buscaCotacaoDolarApiExterna() (*Cotacao, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err().Error() == context.DeadlineExceeded.Error() {
			return nil, context.DeadlineExceeded
		}
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
	log.Println("Cotacao buscada com sucesso: ", cotacao)
	return &cotacao, nil
}

func salvarCotacaoBancoDeDados(db *gorm.DB, cotacao *Cotacao) error {
	log.Println("Salvando a cotação do ativo: ", cotacao.Usdbrl.Name, " no valor de: ", cotacao.Usdbrl.Bid)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	db.WithContext(ctx).Create(&cotacaoEntity{
		Code:       cotacao.Usdbrl.Code,
		Codein:     cotacao.Usdbrl.Codein,
		Name:       cotacao.Usdbrl.Name,
		High:       cotacao.Usdbrl.High,
		Low:        cotacao.Usdbrl.Low,
		VarBid:     cotacao.Usdbrl.VarBid,
		PctChange:  cotacao.Usdbrl.PctChange,
		Bid:        cotacao.Usdbrl.Bid,
		Ask:        cotacao.Usdbrl.Ask,
		Timestamp:  cotacao.Usdbrl.Timestamp,
		CreateDate: cotacao.Usdbrl.CreateDate,
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		log.Println("cotação do ativo: ", cotacao.Usdbrl.Name, " no valor de: ", cotacao.Usdbrl.Bid, " foi salva com sucesso!")
		return nil
	}
}
