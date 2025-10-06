package ticktacktoe

import (
	"fmt"
	"strings"
)

type GameState struct {
	// board holds 9 squares (indexes 0..8) using 'X', 'O' or 0 for empty.
	board [9]rune
	// moves is the chronological sequence of square indices that have been played.
	moves []int
	// winner is 'X' or 'O' once a player has won; 0 means no winner yet.
	winner rune
	// draw is true if the game ended with no winner.
	draw bool
}

func NewGameState() *GameState {
	return &GameState{}
}

var squareOrder = []rune{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I'}
var squareIndex = map[rune]int{
	'A': 0, 'B': 1, 'C': 2,
	'D': 3, 'E': 4, 'F': 5,
	'G': 6, 'H': 7, 'I': 8,
}

var winLines = [8][3]int{
	{0, 1, 2}, // rows
	{3, 4, 5},
	{6, 7, 8},
	{0, 3, 6}, // cols
	{1, 4, 7},
	{2, 5, 8},
	{0, 4, 8}, // diagonals
	{2, 4, 6},
}

// GridToSquare converts a grid address like "B2" (column A-C + row 1-3)
// to the canonical square letter string ("A"-"I").
func GridToSquare(addr string) (string, error) {
	if len(addr) != 2 {
		return "", fmt.Errorf("grid address must be length 2, got %q", addr)
	}
	col := addr[0]
	rowChar := addr[1]
	if col < 'A' || col > 'C' {
		return "", fmt.Errorf("column must be A-C, got %c", col)
	}
	if rowChar < '1' || rowChar > '3' {
		return "", fmt.Errorf("row must be 1-3, got %c", rowChar)
	}
	row := int(rowChar - '1') // 0..2
	rBase := row * 3
	cOffset := int(col - 'A')
	idx := rBase + cOffset
	return string(squareOrder[idx]), nil
}

// SquareToGrid converts a canonical square letter ("A"-"I") to a grid address
// like "B2". Returns error if square invalid.
func SquareToGrid(square string) (string, error) {
	if len(square) != 1 {
		return "", fmt.Errorf("square must be a single letter, got %q", square)
	}
	r := rune(square[0])
	idx, ok := squareIndex[r]
	if !ok {
		return "", fmt.Errorf("invalid square %c", r)
	}
	row := (idx / 3) + 1
	col := rune('A' + (idx % 3))
	return fmt.Sprintf("%c%d", col, row), nil
}

// PlayerToMove returns 'X' or 'O' depending on whose turn it is, or 0 if game over.
func (gs *GameState) PlayerToMove() rune {
	if gs.winner != 0 || gs.draw {
		return 0
	}
	if len(gs.moves)%2 == 0 {
		return 'X'
	}
	return 'O'
}

// GameStateFromString parses a string representation of the game state
// and returns a GameState object. The input string is expected to be
// a sequence of moves as described in the ToString method.
func GameStateFromString(s string) (*GameState, error) {
	gs := &GameState{}
	for i, ch := range s {
		idx, ok := squareIndex[ch]
		if !ok {
			return nil, fmt.Errorf("invalid square '%c' at position %d", ch, i)
		}
		if gs.board[idx] != 0 {
			return nil, fmt.Errorf("square '%c' already occupied (position %d)", ch, i)
		}
		player := gs.PlayerToMove()
		if player == 0 { // game already over but more moves supplied
			return nil, fmt.Errorf("move after game end at position %d", i)
		}
		gs.board[idx] = player
		gs.moves = append(gs.moves, idx)
		gs.updateTerminalState()
	}
	return gs, nil
}

// ToString returns a string representation of the sequence
// of moves in the game state. The game board is understood
// as a set of squares labeled A through I as follows:
//
//	A	B	C
//	D	E	F
//	G	H	I
//
// The player X always moves first, followed by O, and so on.
// The string representation is a sequence of these square
// labels, e.g. "AEI" means X moved to A, O moved to E, and
// X moved to I.
func (gs *GameState) ToString() string {
	bytes := make([]rune, 0, len(gs.moves))
	for _, idx := range gs.moves {
		bytes = append(bytes, squareOrder[idx])
	}
	return string(bytes)
}

func (gs *GameState) ListValidMoves() []string {
	if gs.winner != 0 || gs.draw {
		return nil
	}
	player := gs.PlayerToMove()
	_ = player // player is not used but conceptually relevant; keep to show intent
	moves := make([]string, 0, 9-len(gs.moves))
	for i, r := range gs.board {
		if r == 0 {
			moves = append(moves, string(squareOrder[i]))
		}
	}
	return moves
}

func (gs *GameState) ApplyMove(move string) error {
	if len(move) != 1 {
		return fmt.Errorf("move must be a single square letter A-I")
	}
	if gs.winner != 0 || gs.draw {
		return fmt.Errorf("game already finished")
	}
	ch := rune(move[0])
	idx, ok := squareIndex[ch]
	if !ok {
		return fmt.Errorf("invalid square '%s'", move)
	}
	if gs.board[idx] != 0 {
		return fmt.Errorf("square '%s' already occupied", move)
	}
	player := gs.PlayerToMove()
	if player == 0 {
		return fmt.Errorf("no player to move")
	}
	gs.board[idx] = player
	gs.moves = append(gs.moves, idx)
	gs.updateTerminalState()
	return nil
}

// updateTerminalState updates winner/draw flags after a move or parsing.
func (gs *GameState) updateTerminalState() {
	if gs.winner != 0 || gs.draw {
		return
	}
	// Check win
	for _, line := range winLines {
		a, b, c := line[0], line[1], line[2]
		if gs.board[a] != 0 && gs.board[a] == gs.board[b] && gs.board[a] == gs.board[c] {
			gs.winner = gs.board[a]
			return
		}
	}
	// Check draw
	if len(gs.moves) == 9 {
		gs.draw = true
	}
}

// Winner returns 'X', 'O', or 0 if no winner.
func (gs *GameState) Winner() rune { return gs.winner }

// IsDraw returns true if the game ended in a draw.
func (gs *GameState) IsDraw() bool { return gs.draw }

// BoardString returns a multi-line human-readable board representation including
// column headers (A-C) and row numbers (1-3). Example (final drawn position):
//
//	  A   B   C
//	+---+---+---+
//
// 1 | X | O | X |
//
//	+---+---+---+
//
// 2 | X | X | O |
//
//	+---+---+---+
//
// 3 | O | X | O |
//
//	+---+---+---+
//
// Empty squares render as a single space. A trailing newline is included to
// ease direct printing. This format is stable for snapshot tests.
func (gs *GameState) BoardString() string {
	var b strings.Builder
	// Column header
	b.WriteString("    A   B   C\n")
	b.WriteString("  +---+---+---+\n")
	for r := 0; r < 3; r++ {
		b.WriteString(fmt.Sprintf("%d |", r+1))
		for c := 0; c < 3; c++ {
			idx := r*3 + c
			ch := gs.board[idx]
			if ch == 0 {
				ch = ' '
			}
			b.WriteString(" ")
			b.WriteRune(ch)
			b.WriteString(" |")
		}
		b.WriteByte('\n')
		b.WriteString("  +---+---+---+\n")
	}
	return b.String()
}
