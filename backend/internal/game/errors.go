package game

import "errors"

var (
	ErrGameNotFound      = errors.New("game not found")
	ErrGameNotInProgress = errors.New("game not in progress")
	ErrInvalidPlayer     = errors.New("invalid player")
	ErrNotYourTurn       = errors.New("not your turn")
	ErrInvalidMove       = errors.New("invalid move")
	ErrColumnFull        = errors.New("column is full")
)
