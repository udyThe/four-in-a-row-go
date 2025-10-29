package game

import (
	"errors"
	"fmt"
)

const (
	Rows    = 6
	Columns = 7
)

type CellState int

const (
	Empty CellState = iota
	Player1
	Player2
)

type Board struct {
	Grid [Rows][Columns]CellState
}

func NewBoard() *Board {
	return &Board{}
}

// DropDisc drops a disc into the specified column
func (b *Board) DropDisc(column int, player CellState) (int, error) {
	if column < 0 || column >= Columns {
		return -1, errors.New("invalid column")
	}

	// Find the lowest available row in the column
	for row := Rows - 1; row >= 0; row-- {
		if b.Grid[row][column] == Empty {
			b.Grid[row][column] = player
			return row, nil
		}
	}

	return -1, errors.New("column is full")
}

// IsValidMove checks if a move is valid
func (b *Board) IsValidMove(column int) bool {
	if column < 0 || column >= Columns {
		return false
	}
	return b.Grid[0][column] == Empty
}

// GetValidMoves returns all valid column indices
func (b *Board) GetValidMoves() []int {
	var moves []int
	for col := 0; col < Columns; col++ {
		if b.IsValidMove(col) {
			moves = append(moves, col)
		}
	}
	return moves
}

// CheckWin checks if the specified player has won
func (b *Board) CheckWin(player CellState) bool {
	// Check horizontal
	for row := 0; row < Rows; row++ {
		for col := 0; col <= Columns-4; col++ {
			if b.Grid[row][col] == player &&
				b.Grid[row][col+1] == player &&
				b.Grid[row][col+2] == player &&
				b.Grid[row][col+3] == player {
				return true
			}
		}
	}

	// Check vertical
	for row := 0; row <= Rows-4; row++ {
		for col := 0; col < Columns; col++ {
			if b.Grid[row][col] == player &&
				b.Grid[row+1][col] == player &&
				b.Grid[row+2][col] == player &&
				b.Grid[row+3][col] == player {
				return true
			}
		}
	}

	// Check diagonal (bottom-left to top-right)
	for row := 3; row < Rows; row++ {
		for col := 0; col <= Columns-4; col++ {
			if b.Grid[row][col] == player &&
				b.Grid[row-1][col+1] == player &&
				b.Grid[row-2][col+2] == player &&
				b.Grid[row-3][col+3] == player {
				return true
			}
		}
	}

	// Check diagonal (top-left to bottom-right)
	for row := 0; row <= Rows-4; row++ {
		for col := 0; col <= Columns-4; col++ {
			if b.Grid[row][col] == player &&
				b.Grid[row+1][col+1] == player &&
				b.Grid[row+2][col+2] == player &&
				b.Grid[row+3][col+3] == player {
				return true
			}
		}
	}

	return false
}

// IsFull checks if the board is completely filled
func (b *Board) IsFull() bool {
	for col := 0; col < Columns; col++ {
		if b.Grid[0][col] == Empty {
			return false
		}
	}
	return true
}

// Copy creates a deep copy of the board
func (b *Board) Copy() *Board {
	newBoard := &Board{}
	for row := 0; row < Rows; row++ {
		for col := 0; col < Columns; col++ {
			newBoard.Grid[row][col] = b.Grid[row][col]
		}
	}
	return newBoard
}

// String returns a string representation of the board
func (b *Board) String() string {
	result := ""
	for row := 0; row < Rows; row++ {
		for col := 0; col < Columns; col++ {
			switch b.Grid[row][col] {
			case Empty:
				result += ". "
			case Player1:
				result += "X "
			case Player2:
				result += "O "
			}
		}
		result += "\n"
	}
	return result
}

// ToArray converts the board to a 2D array for JSON serialization
func (b *Board) ToArray() [][]int {
	arr := make([][]int, Rows)
	for row := 0; row < Rows; row++ {
		arr[row] = make([]int, Columns)
		for col := 0; col < Columns; col++ {
			arr[row][col] = int(b.Grid[row][col])
		}
	}
	return arr
}

// FromArray loads board state from a 2D array
func (b *Board) FromArray(arr [][]int) error {
	if len(arr) != Rows {
		return fmt.Errorf("invalid board height: expected %d, got %d", Rows, len(arr))
	}
	
	for row := 0; row < Rows; row++ {
		if len(arr[row]) != Columns {
			return fmt.Errorf("invalid board width at row %d: expected %d, got %d", row, Columns, len(arr[row]))
		}
		for col := 0; col < Columns; col++ {
			b.Grid[row][col] = CellState(arr[row][col])
		}
	}
	return nil
}
