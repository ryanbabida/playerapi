package main

import (
	"fmt"
	"math"
	"net/http"
)

type players interface {
	GetPlayers(name *string, minAge *int, maxAge *int, position *string, sport *string) ([]Player, error)
	GetAverageAgeByPosition() (map[string]float64, error)
}

type Service struct {
	players players
}

type ServiceResponse[T any] struct {
	Data       T `json:"data"`
	statusCode int
	Error      error `json:"error"`
}

type GetPlayersQuery struct {
	Name     *string `json:"name"`
	MinAge   *int    `json:"minAge"`
	MaxAge   *int    `json:"maxAge"`
	Position *string `json:"position"`
	Sport    *string `json:"sport"`
}

func (s *Service) GetPlayers(query GetPlayersQuery) ServiceResponse[[]PlayerResponse] {
	result := []PlayerResponse{}

	players, err := s.players.GetPlayers(query.Name, query.MinAge, query.MaxAge, query.Position, query.Sport)
	if err != nil {
		return ServiceResponse[[]PlayerResponse]{Data: nil, statusCode: http.StatusInternalServerError, Error: err}
	}

	avgAges, err := s.players.GetAverageAgeByPosition()
	if err != nil {
		return ServiceResponse[[]PlayerResponse]{Data: nil, statusCode: http.StatusInternalServerError, Error: err}
	}

	for _, player := range players {
		avgAge, ok := avgAges[player.Position]
		if !ok {
			continue
		}

		name, err := getNameBrief(player.FirstName, player.LastName, player.Sport)
		if err != nil {
			continue
		}

		result = append(result, PlayerResponse{
			Id:                     player.Id,
			NameBrief:              name,
			FirstName:              player.FirstName,
			LastName:               player.LastName,
			Position:               player.Position,
			Age:                    player.Age,
			AveragePositionAgeDiff: player.Age - int(math.Round(avgAge)),
		})
	}

	return ServiceResponse[[]PlayerResponse]{Data: result, statusCode: http.StatusOK, Error: nil}
}

func getNameBrief(firstName string, lastName string, sport string) (string, error) {
	switch sport {
	case "football":
		return fmt.Sprintf("%c. %s", firstName[0], lastName), nil
	case "basketball":
		return fmt.Sprintf("%s %c.", firstName, lastName[0]), nil
	case "baseball":
		return fmt.Sprintf("%c. %c.", firstName[0], lastName[0]), nil
	}

	return "", fmt.Errorf("unknown sport '%s' found", sport)
}
