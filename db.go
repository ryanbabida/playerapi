package main

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type Player struct {
	Id        int    `db:"id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Position  string `db:"position"`
	Age       int    `db:"age"`
	Sport     string `db:"sport"`
}

type playerStore struct {
	db *sqlx.DB
}

func NewPlayerStore(db *sqlx.DB) *playerStore {
	return &playerStore{db}
}

func (p *playerStore) GetPlayers(name *string, minAge *int, maxAge *int, position *string, sport *string) ([]Player, error) {
	players := []Player{}
	queryBuilder := sq.Select("*").
		From("players")

	if name != nil {
		queryBuilder = queryBuilder.
			Where(sq.Or{
				sq.Expr("lower(first_name) like lower(?)", "%"+*name+"%"),
				sq.Expr("lower(last_name) like lower(?)", "%"+*name+"%"),
			})
	}

	if minAge != nil {
		queryBuilder = queryBuilder.Where(sq.GtOrEq{"age": minAge})
	}

	if maxAge != nil {
		queryBuilder = queryBuilder.Where(sq.LtOrEq{"age": maxAge})
	}

	if position != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"position": strings.ToLower(*position)})
	}

	if sport != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"sport": strings.ToLower(*sport)})
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed create select query: %v", err)
	}

	err = p.db.Select(&players, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to read from players table: %v", err)
	}

	return players, nil
}

type averageAgeByPosition struct {
	Position string  `db:"position"`
	Age      float64 `db:"age"`
}

func (p *playerStore) GetAverageAgeByPosition() (map[string]float64, error) {
	avgAges := []averageAgeByPosition{}
	queryBuilder := sq.Select("position", "avg(age) as age").
		From("players").
		Where(sq.Gt{"age": 0}).
		GroupBy("position")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed create avg age query: %v", err)
	}

	err = p.db.Select(&avgAges, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve average age by position from player table: %v", err)
	}

	avgAgeByPos := map[string]float64{}
	for _, row := range avgAges {
		avgAgeByPos[row.Position] = row.Age
	}

	return avgAgeByPos, nil
}
