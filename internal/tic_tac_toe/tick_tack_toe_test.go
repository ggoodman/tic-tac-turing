package ticktacktoe

import "testing"

func TestParseAndSerializeEmpty(t *testing.T) {
	gs, err := GameStateFromString("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gs.ToString() != "" {
		t.Fatalf("expected empty serialization, got %q", gs.ToString())
	}
	if gs.PlayerToMove() != 'X' {
		t.Fatalf("expected X to move first")
	}
}

func TestApplyMovesAndSerialize(t *testing.T) {
	gs, _ := GameStateFromString("")
	moves := []string{"A", "E", "B"}
	for _, m := range moves {
		if err := gs.ApplyMove(m); err != nil {
			t.Fatalf("apply %s: %v", m, err)
		}
	}
	if got := gs.ToString(); got != "AEB" {
		t.Fatalf("expected AEB got %s", got)
	}
	if gs.PlayerToMove() != 'O' {
		t.Fatalf("expected O to move next")
	}
}

func TestWinDetectionRow(t *testing.T) {
	gs, _ := GameStateFromString("AB") // X:A, O:B
	if err := gs.ApplyMove("D"); err != nil {
		t.Fatal(err)
	} // X:D
	if err := gs.ApplyMove("E"); err != nil {
		t.Fatal(err)
	} // O:E
	if err := gs.ApplyMove("G"); err != nil {
		t.Fatal(err)
	} // X:G wins first column
	if gs.Winner() != 'X' {
		t.Fatalf("expected X winner, got %c", gs.Winner())
	}
	if err := gs.ApplyMove("C"); err == nil {
		t.Fatalf("expected error applying move after game end")
	}
}

func TestInvalidDuplicateSquare(t *testing.T) {
	if _, err := GameStateFromString("AA"); err == nil {
		t.Fatalf("expected error for duplicate square")
	}
}

func TestDraw(t *testing.T) {
	// Sequence leading to draw producing board:
	// X O X
	// X X O
	// O X O
	// Order: X:A O:B X:C O:F X:D O:G X:E O:I X:H
	seq := "ABCFDGEIH"
	gs, err := GameStateFromString(seq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gs.Winner() != 0 {
		t.Fatalf("expected no winner got %c", gs.Winner())
	}
	if !gs.IsDraw() {
		t.Fatalf("expected draw")
	}
}

func TestListValidMoves(t *testing.T) {
	gs, _ := GameStateFromString("A")
	moves := gs.ListValidMoves()
	foundB := false
	for _, m := range moves {
		if m == "B" {
			foundB = true
		}
	}
	if !foundB {
		t.Fatalf("expected B in valid moves")
	}
	for _, m := range moves {
		if m == "A" {
			t.Fatalf("A should not be valid")
		}
	}
}

func TestBoardString(t *testing.T) {
	gs, _ := GameStateFromString("ABCFDGEIH") // draw board
	expected := "" +
		"    A   B   C\n" +
		"  +---+---+---+\n" +
		"1 | X | O | X |\n" +
		"  +---+---+---+\n" +
		"2 | X | X | O |\n" +
		"  +---+---+---+\n" +
		"3 | O | X | O |\n" +
		"  +---+---+---+\n"
	if got := gs.BoardString(); got != expected {
		t.Fatalf("BoardString mismatch.\nExpected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestGridAddressTranslation(t *testing.T) {
	cases := []struct{ grid, square string }{
		{"A1", "A"}, {"B1", "B"}, {"C1", "C"},
		{"A2", "D"}, {"B2", "E"}, {"C2", "F"},
		{"A3", "G"}, {"B3", "H"}, {"C3", "I"},
	}
	for _, tc := range cases {
		sq, err := GridToSquare(tc.grid)
		if err != nil {
			t.Fatalf("grid %s: %v", tc.grid, err)
		}
		if sq != tc.square {
			t.Fatalf("expected %s got %s", tc.square, sq)
		}
		grid, err := SquareToGrid(tc.square)
		if err != nil {
			t.Fatalf("square %s: %v", tc.square, err)
		}
		if grid != tc.grid {
			t.Fatalf("reverse expected %s got %s", tc.grid, grid)
		}
	}
}
