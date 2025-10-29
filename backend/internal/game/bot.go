package game

import (
	"math"
	"math/rand"
	"time"
)

const (
	MaxDepth    = 6 // Maximum search depth for minimax
	WinScore    = 1000000
	ThreeScore  = 100
	TwoScore    = 10
	CenterBonus = 3
)

type Bot struct {
	player   CellState
	opponent CellState
	rand     *rand.Rand
}

func NewBot(player CellState) *Bot {
	opponent := Player1
	if player == Player1 {
		opponent = Player2
	}
	return &Bot{
		player:   player,
		opponent: opponent,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetBestMove returns the best column to play using minimax with alpha-beta pruning
func (bot *Bot) GetBestMove(board *Board) int {
	validMoves := board.GetValidMoves()
	if len(validMoves) == 0 {
		return -1
	}

	// Check for immediate winning move
	for _, col := range validMoves {
		testBoard := board.Copy()
		testBoard.DropDisc(col, bot.player)
		if testBoard.CheckWin(bot.player) {
			return col
		}
	}

	// Check for blocking opponent's winning move
	for _, col := range validMoves {
		testBoard := board.Copy()
		testBoard.DropDisc(col, bot.opponent)
		if testBoard.CheckWin(bot.opponent) {
			return col
		}
	}

	// Use minimax to find the best move
	bestScore := math.Inf(-1)
	bestMoves := []int{}

	for _, col := range validMoves {
		testBoard := board.Copy()
		row, _ := testBoard.DropDisc(col, bot.player)
		
		// Evaluate immediate position
		score := float64(bot.evaluateWindow(testBoard, row, col))
		
		// Add minimax score
		score += bot.minimax(testBoard, MaxDepth-1, math.Inf(-1), math.Inf(1), false)

		if score > bestScore {
			bestScore = score
			bestMoves = []int{col}
		} else if score == bestScore {
			bestMoves = append(bestMoves, col)
		}
	}

	// Return random best move if multiple exist
	if len(bestMoves) > 0 {
		return bestMoves[bot.rand.Intn(len(bestMoves))]
	}

	return validMoves[0]
}

// minimax implements the minimax algorithm with alpha-beta pruning
func (bot *Bot) minimax(board *Board, depth int, alpha, beta float64, maximizing bool) float64 {
	// Terminal conditions
	if board.CheckWin(bot.player) {
		return WinScore + float64(depth)
	}
	if board.CheckWin(bot.opponent) {
		return -WinScore - float64(depth)
	}
	if board.IsFull() || depth == 0 {
		return bot.evaluateBoard(board)
	}

	validMoves := board.GetValidMoves()

	if maximizing {
		maxEval := math.Inf(-1)
		for _, col := range validMoves {
			testBoard := board.Copy()
			testBoard.DropDisc(col, bot.player)
			eval := bot.minimax(testBoard, depth-1, alpha, beta, false)
			maxEval = math.Max(maxEval, eval)
			alpha = math.Max(alpha, eval)
			if beta <= alpha {
				break // Beta cutoff
			}
		}
		return maxEval
	} else {
		minEval := math.Inf(1)
		for _, col := range validMoves {
			testBoard := board.Copy()
			testBoard.DropDisc(col, bot.opponent)
			eval := bot.minimax(testBoard, depth-1, alpha, beta, true)
			minEval = math.Min(minEval, eval)
			beta = math.Min(beta, eval)
			if beta <= alpha {
				break // Alpha cutoff
			}
		}
		return minEval
	}
}

// evaluateBoard scores the entire board position
func (bot *Bot) evaluateBoard(board *Board) float64 {
	score := 0.0

	// Center column preference
	centerCol := Columns / 2
	centerCount := 0
	for row := 0; row < Rows; row++ {
		if board.Grid[row][centerCol] == bot.player {
			centerCount++
		}
	}
	score += float64(centerCount * CenterBonus)

	// Evaluate all possible windows
	score += bot.evaluateAllWindows(board)

	return score
}

// evaluateAllWindows checks all 4-cell windows for potential threats/opportunities
func (bot *Bot) evaluateAllWindows(board *Board) float64 {
	score := 0.0

	// Horizontal windows
	for row := 0; row < Rows; row++ {
		for col := 0; col <= Columns-4; col++ {
			window := []CellState{
				board.Grid[row][col],
				board.Grid[row][col+1],
				board.Grid[row][col+2],
				board.Grid[row][col+3],
			}
			score += bot.scoreWindow(window)
		}
	}

	// Vertical windows
	for row := 0; row <= Rows-4; row++ {
		for col := 0; col < Columns; col++ {
			window := []CellState{
				board.Grid[row][col],
				board.Grid[row+1][col],
				board.Grid[row+2][col],
				board.Grid[row+3][col],
			}
			score += bot.scoreWindow(window)
		}
	}

	// Diagonal windows (bottom-left to top-right)
	for row := 3; row < Rows; row++ {
		for col := 0; col <= Columns-4; col++ {
			window := []CellState{
				board.Grid[row][col],
				board.Grid[row-1][col+1],
				board.Grid[row-2][col+2],
				board.Grid[row-3][col+3],
			}
			score += bot.scoreWindow(window)
		}
	}

	// Diagonal windows (top-left to bottom-right)
	for row := 0; row <= Rows-4; row++ {
		for col := 0; col <= Columns-4; col++ {
			window := []CellState{
				board.Grid[row][col],
				board.Grid[row+1][col+1],
				board.Grid[row+2][col+2],
				board.Grid[row+3][col+3],
			}
			score += bot.scoreWindow(window)
		}
	}

	return score
}

// scoreWindow evaluates a 4-cell window
func (bot *Bot) scoreWindow(window []CellState) float64 {
	botCount := 0
	oppCount := 0
	emptyCount := 0

	for _, cell := range window {
		if cell == bot.player {
			botCount++
		} else if cell == bot.opponent {
			oppCount++
		} else {
			emptyCount++
		}
	}

	// Can't use window if both players have pieces
	if botCount > 0 && oppCount > 0 {
		return 0
	}

	// Score bot's opportunities
	if botCount == 4 {
		return WinScore
	} else if botCount == 3 && emptyCount == 1 {
		return ThreeScore
	} else if botCount == 2 && emptyCount == 2 {
		return TwoScore
	}

	// Penalize opponent's opportunities
	if oppCount == 3 && emptyCount == 1 {
		return -ThreeScore * 1.5 // Slightly prioritize blocking
	} else if oppCount == 2 && emptyCount == 2 {
		return -TwoScore
	}

	return 0
}

// evaluateWindow scores a specific position
func (bot *Bot) evaluateWindow(board *Board, row, col int) float64 {
	score := 0.0

	// Check all directions from this position
	directions := [][2]int{
		{0, 1},  // Horizontal
		{1, 0},  // Vertical
		{1, 1},  // Diagonal \
		{1, -1}, // Diagonal /
	}

	for _, dir := range directions {
		window := bot.getWindow(board, row, col, dir[0], dir[1])
		if len(window) == 4 {
			score += bot.scoreWindow(window)
		}
	}

	return score
}

// getWindow extracts a 4-cell window in a given direction
func (bot *Bot) getWindow(board *Board, startRow, startCol, dRow, dCol int) []CellState {
	var window []CellState
	for i := 0; i < 4; i++ {
		row := startRow + i*dRow
		col := startCol + i*dCol
		if row >= 0 && row < Rows && col >= 0 && col < Columns {
			window = append(window, board.Grid[row][col])
		}
	}
	return window
}
