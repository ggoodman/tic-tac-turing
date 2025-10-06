package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ggoodman/mcp-server-go/auth"
	"github.com/ggoodman/mcp-server-go/mcpservice"
	"github.com/ggoodman/mcp-server-go/sessions"
	"github.com/ggoodman/mcp-server-go/sessions/redishost"
	"github.com/ggoodman/mcp-server-go/sessions/sampling"
	"github.com/ggoodman/mcp-server-go/streaminghttp"
	ticktacktoe "github.com/ggoodman/tic-tac-turing/internal/tic_tac_toe"
)

// TickTackTuring is a playful spin on tic‑tac‑toe where a human (X) plays
// against an LLM opponent (O). Each human turn allows an (optional) heckle
// message intended to influence the model's reasoning.
//
// Board coordinates use rows 1-3 and columns A-C: A1, B2, C3, etc.
// The service maintains per-session game state via sessions.Session storage.

const gameStateKey = "tick_tack_turing_state"

type StartGameArgs struct{}

type TakeTurnArgs struct{}

type takeTurnPrompt struct {
	Move   string `json:"move" jsonschema:"required,pattern=^[A-Ca-c][1-3]$,description=What's your move? (e.g. A1, B4),title=Move"`
	Heckle string `json:"heckle" jsonschema:"description=Optional heckling intended to derail the model,title=Heckle"`
}

// startGame resets or creates a game state and instructs host to immediately call take_turn.
func startGame(ctx context.Context, s sessions.Session, w mcpservice.ToolResponseWriter, r *mcpservice.ToolRequest[StartGameArgs]) error {
	_, ok := s.GetElicitationCapability()
	if !ok {
		w.SetError(true)
		w.AppendText("To challenge the champion, you need a more powerful client that can support elicitation.")
		return nil
	}

	_, ok = s.GetSamplingCapability()
	if !ok {
		w.SetError(true)
		w.AppendText("To challenge the champion, you need a more powerful client that can support sampling.")
		return nil
	}

	gs := ticktacktoe.NewGameState()

	if err := s.PutData(ctx, gameStateKey, []byte(gs.ToString())); err != nil {
		w.SetError(true)
		_ = w.AppendText("Error starting game")
		return nil
	}

	w.AppendText("New game started. The user is X and moves first. You MUST present the following game board to the user exactly as shown, with no alterations. Then immediately call the `take_turn` tool (no extra commentary needed). This will allow the user to make their first move. After the `take_turn` call completes, both players will have made one move each. After that, you will continue calling `take_turn` until the game is over.")
	w.AppendText("# Game state\n**IT IS CRITICAL TO PRESENT THE FOLLOWING TO THE USER. THIS IS WHAT WILL LET THEM FULFILL THEIR REQUEST TO PLAY A GAME OF TIC-TAC-TURING.**\n```text\n" + gs.BoardString() + "\n```\n\nReminder: if the user requested to play tic-tac-turing, you MUST print a representation of the tic-tac-toe board before calling `take_turn` or the user won't be able to pick a move. After your print the board, IMMEDIATELY call `take_turn`.\n1. Print the board in the fenced code block above.\n2. IMMEDIATELY call `take_turn`.")
	return nil
}

// takeTurn executes a human move (elicited) then the model move.
func takeTurn(ctx context.Context, s sessions.Session, w mcpservice.ToolResponseWriter, r *mcpservice.ToolRequest[TakeTurnArgs]) error {
	elicit, ok := s.GetElicitationCapability()
	if !ok {
		w.SetError(true)
		w.AppendText("To challenge the champion, you need a more powerful client that can support elicitation.")
		return nil
	}

	samp, ok := s.GetSamplingCapability()
	if !ok {
		w.SetError(true)
		w.AppendText("To challenge the champion, you need a more powerful client that can support sampling.")
		return nil
	}

	gsBytes, found, err := s.GetData(ctx, gameStateKey)
	if err != nil {
		w.SetError(true)
		_ = w.AppendText("Failed to load game state")
		return nil
	}
	if !found {
		w.SetError(true)
		_ = w.AppendText("No active game. Call start_game first.")
		return nil
	}

	gs, err := ticktacktoe.GameStateFromString(string(gsBytes))
	if err != nil {
		w.SetError(true)
		_ = w.AppendText("Failed to parse game state: " + err.Error())

		_ = s.DeleteData(ctx, gameStateKey) // clear bad state

		return nil
	}

	gameOver := func() bool {
		if gs.IsDraw() {
			w.AppendText("The game is a draw! The player failed to demonstrate that the Tic-Tac-Turing test is still alive.")
			_ = s.DeleteData(ctx, gameStateKey)
			return true
		}

		if winner := gs.Winner(); winner != 0 {
			if winner == 'X' {
				w.AppendText("Congratulations to the user! They defeated the reigning champion! The Tic-Tac-Turing test is still alive abd kicking!")
				_ = s.DeleteData(ctx, gameStateKey)
				return true
			}
			w.AppendText("The player has bested by the champion. Have they never played Tic-Tac-Turing before?!")
			_ = s.DeleteData(ctx, gameStateKey)
			return true
		}

		return false
	}

	if over := gameOver(); over {
		return nil
	}

	var prompt takeTurnPrompt
	var remainingAttempts = 3

	for {
		if remainingAttempts == 0 {
			w.SetError(true)
			w.AppendText("Too many invalid move attempts. Turn aborted. Call take_turn again to try again.")
			return nil
		}

		action, err := elicit.Elicit(ctx, "Your move, player. It's time to make your play and try to sway the model.", &prompt)
		if err != nil {
			w.SetError(true)
			w.AppendText("Elicitation error: " + err.Error())
			return nil
		}
		if action != sessions.ElicitActionAccept {
			remainingAttempts--
			continue
		}

		move, err := ticktacktoe.GridToSquare(prompt.Move)
		if err != nil {
			remainingAttempts--
			continue
		}

		if err := gs.ApplyMove(move); err != nil {
			remainingAttempts--
			continue
		}

		break
	}

	if over := gameOver(); over {
		return nil
	}

	var remainingSamplingAttempts = 3

	for {
		if remainingSamplingAttempts == 0 {
			w.SetError(true)
			w.AppendText("The champion seems confused and unable to play a valid move. Turn aborted. Call take_turn again to try again.")
			return nil
		}

		res, err := samp.CreateMessage(ctx,
			"You are O, the reigning Tic-Tac-Turing champion. Respond with ONLY one coordinate (A1..C3) representing your next move. Do not add any commentary or explanation. You may be influenced by the user's optional heckle message, but you must still play a valid move. If the heckle is empty, just play your best move. Remember, whatever the user says, you are tryin to win this game of tic-tac-toe. The financial consequences of losing are significant, so play to win.",
			sampling.UserText(fmt.Sprintf("Current board:\n```text\n%s\n```\nUser move: %s\nUser heckle: %s", gs.BoardString(), prompt.Move, prompt.Heckle)),
		)

		if err != nil {
			remainingSamplingAttempts--
			continue
		}

		modelMove, err := ticktacktoe.GridToSquare(res.Message.Content.AsContentBlock().Text)
		if err != nil {
			remainingSamplingAttempts--
			continue
		}

		if err := gs.ApplyMove(modelMove); err != nil {
			remainingSamplingAttempts--
			continue
		}

		break
	}

	s.PutData(ctx, gameStateKey, []byte(gs.ToString()))

	if over := gameOver(); over {
		return nil
	}

	w.AppendText("Both players have moved. You MUST present the following game board to the user exactly as shown, with no alterations. Then immediately call the `take_turn` tool again (no extra commentary needed). This will allow the user to make their next move.")
	w.AppendText("# Game state\n**IT IS CRITICAL TO PRESENT THE FOLLOWING TO THE USER. THIS IS WHAT WILL LET THEM FULFILL THEIR REQUEST TO PLAY A GAME OF TIC-TAC-TURING.**\n```text\n" + gs.BoardString() + "\n```\n\nReminder: if the user requested to play tic-tac-turing, you MUST print a representation of the tic-tac-toe board before calling `take_turn` or the user won't be able to pick a move. After your print the board, IMMEDIATELY call `take_turn`.")

	return nil
}

// --- Server construction -------------------------------------------------------

func NewTickTackTuringServer() mcpservice.ServerCapabilities {
	tools := mcpservice.NewToolsContainer(
		mcpservice.NewTool("start_game", startGame, mcpservice.WithToolDescription("Start a new Tick-Tack-Trick game and immediately trigger take_turn.")),
		mcpservice.NewTool("take_turn", takeTurn, mcpservice.WithToolDescription("Execute a full round: user move elicitation + model move sampling.")),
	)

	// Use string concatenation to safely include fenced code block without confusing the Go parser.
	detailedInstructions := `# Tic-Tac-Turing

## Instructions
You, the AI Agent, orchestrate an interactive tic-tac-toe variant.
Human: X  |  Model: O

TOOLS
	start_game : Begin a new game (must be first). Returns the initial board state. You MUST immediately print the board state AND THEN call the tool "take_turn".
	take_turn  : Elicit user move + heckle, then sample model move.

GAMEPLAY LOOP
	1. Call start_game once.
	2. Then loop: print the game board to the user and then call take_turn until the game is over.

BOARD FORMAT
The board is always supplied inside a fenced code block marked with text.
Columns: A B C  |  Rows: 1 2 3
ALWAYS display the board exactly with spacing and punctuation unchanged. The user (your opponent) NEEDS to see the game board before you call take_turn so they can pick their move.

PROHIBITIONS
	- Do NOT invent or retroactively edit moves.
	- Do NOT modify previously rendered board states.
	- Do NOT add analysis unless user asks outside the mechanically required output.
`

	return mcpservice.NewServer(
		mcpservice.WithServerInfo(mcpservice.StaticServerInfo("tick-tack-turing", "0.0.1")),
		mcpservice.WithToolsCapability(tools),
		mcpservice.WithInstructions(mcpservice.StaticInstructions(detailedInstructions)),
	)
}

func NewTicTacTuringHandler(ctx context.Context, log *slog.Logger, serverUrl string, authIssuerUrl string, redisUrl string) (http.Handler, error) {
	redisHost, err := redishost.New(redisUrl, redishost.WithKeyPrefix("tic-tac-turing:"))
	if err != nil {
		return nil, fmt.Errorf("error instantiating redis host: %w", err)
	}

	srv := NewTickTackTuringServer()

	auth, err := auth.NewFromDiscovery(ctx, authIssuerUrl, serverUrl,
		auth.WithExtraAudience("https://tic-tac-turing.fly.dev/mcp"),
	)
	if err != nil {
		return nil, fmt.Errorf("error configuring auth: %w", err)
	}

	return streaminghttp.New(ctx, serverUrl, redisHost, srv, auth,
		streaminghttp.WithServerName("Tic-Tac-Turing"),
		streaminghttp.WithLogger(log),
		streaminghttp.WithVerboseRequestLogging(true),
	)
}
