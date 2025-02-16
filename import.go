package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type PlayerResponse struct {
	Id                     int    `db:"id" json:"id"`
	NameBrief              string `db:"name_brief" json:"nameBrief"`
	FirstName              string `db:"first_name" json:"firstName"`
	LastName               string `db:"last_name" json:"lastName"`
	Position               string `db:"position" json:"position"`
	Age                    int    `db:"age" json:"age"`
	AveragePositionAgeDiff int    `db:"avg_pos_age_diff" json:"averagePositionAgeDiff"`
}

type CbsPlayersResponse struct {
	Body *Body `json:"body"`
}

type Body struct {
	Players []CbsPlayer `json:"players"`
}

type CbsPlayer struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Position  string `json:"position"`
	Age       int    `db:"age" json:"age"`
}

func initSchema(db *sqlx.DB) {
	schema := `
	create table players (
		id integer primary key autoincrement,
		first_name text not null,
		last_name text not null,
		position integer not null,
		age int not null,
		sport text not null
	);`

	db.MustExec(schema)
}

func initData(db *sqlx.DB, cbsApiUrl string, sports []string, importBatchSize int) int64 {
	var rowCount int64
	for _, sport := range sports {
		fmt.Printf("importing %s data...\n", sport)
		time.Sleep(2 * time.Second)

		url := strings.ReplaceAll(cbsApiUrl, "{{SPORT}}", sport)
		playersResponse, err := getCbsPlayers(url)
		if err != nil {
			log.Fatalln(err)
		}

		if playersResponse == nil || playersResponse.Body == nil {
			log.Fatalf("able to ping api by players response was nil: %s\n", url)
		}

		rowsAffected, err := insertPlayers(db, playersResponse.Body.Players, sport, importBatchSize)
		if err != nil {
			log.Fatalln(err)
		}

		rowCount += rowsAffected
	}

	return rowCount
}

func getCbsPlayers(url string) (*CbsPlayersResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var playersResponse CbsPlayersResponse
	if err := json.Unmarshal(body, &playersResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)

	}

	return &playersResponse, nil
}

func insertPlayers(db *sqlx.DB, players []CbsPlayer, sport string, importBatchSize int) (int64, error) {
	var rowCount int64

	for i := 0; i < len(players); i += importBatchSize {
		end := i + importBatchSize
		if end > len(players) {
			end = len(players)
		}

		count, err := insert(db, players[i:end], sport)
		if err != nil {
			return 0, fmt.Errorf("failed to insert chunk: %v", err)
		}

		rowCount += count
	}

	return rowCount, nil
}

func insert(db *sqlx.DB, players []CbsPlayer, sport string) (int64, error) {
	queryBuilder := sq.Insert("players").Columns("first_name", "last_name", "position", "age", "sport")
	for _, player := range players {
		if player.FirstName != "" {
			queryBuilder = queryBuilder.Values(player.FirstName, player.LastName, player.Position, player.Age, sport)
		}
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to generate sql from parsed response: %v", err)
	}

	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query on players table: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to report rows affected from players table: %v", err)
	}

	return rowsAffected, nil
}
