package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	_bomb    = "☸"
	_player1 = "⚘"
	_player2 = "⚗"
	_dead    = "☠"
	_wall    = "☲"

	timeout    = time.Second * 5
	_bomb_t    = time.Second * 2
	_explode_t = time.Second
)

type bombMsg struct {
	x int
	y int
}

type unfreezeMsg struct {
	P     *player
	bombX int
	bombY int
}

type exploseMsg struct {
	e explosion
}

type board struct {
	Width  int
	Height int
	Arr    [][]int
}

type player struct {
	x          int
	y          int
	freezeBomb bool
}

type bomb struct {
	x int
	y int
}

type explosion [][]int

type game struct {
	P1     *player
	P2     *player
	Board  board
	Ticks  int
	Frames int
}

func initBoard(mapFile string) board {
	f, err := os.Open(mapFile)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		log.Fatal(scanner.Err().Error())
	}

	firstline := strings.Split(scanner.Text(), " ")
	m, n := 0, 0
	m, err = strconv.Atoi(firstline[0])
	if err != nil {
		log.Fatal(err)
	}
	n, err = strconv.Atoi(firstline[1])
	if err != nil {
		log.Fatal(err)
	}

	arr := make([][]int, m)

	x := 0
	for scanner.Scan() {
		row := scanner.Text()
		if row == "" {
			break
		}
		arr[x] = make([]int, n)
		for y, cell := range row {
			arr[x][y] = int(cell) - 48
		}
		x++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return board{
		Width:  n,
		Height: m,
		Arr:    arr,
	}
}

func initGame(filePath string) game {
	board := initBoard(filePath)
	return game{
		P1:    &player{1, 1, false},
		P2:    &player{board.Height - 2, board.Width - 2, false},
		Board: board,
	}
}

func (g game) Init() tea.Cmd { return nil }

func (g game) View() string {
	boardBuilder := make([]string, g.Board.Height)
	for i := 0; i < g.Board.Height; i++ {
		rowBuilder := make([]string, g.Board.Width)
		for j := 0; j < g.Board.Width; j++ {
			if i == g.P1.x && j == g.P1.y {
				rowBuilder[j] = _player1
			} else if i == g.P2.x && j == g.P2.y {
				rowBuilder[j] = _player2
			} else if g.Board.Arr[i][j] == 1 {
				rowBuilder[j] = _wall
			} else if g.Board.Arr[i][j] == 2 {
				rowBuilder[j] = _bomb
			} else if g.Board.Arr[i][j] < 0 {
				rowBuilder[j] = "*"
			} else {
				rowBuilder[j] = " "
			}
		}
		boardBuilder[i] = strings.Join(rowBuilder, " ")
	}
	return strings.Join(boardBuilder, "\n")
}

func (g game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	player := g.P1
	switch msg.(type) {
	case tea.KeyMsg:
		switch msg := msg.(tea.KeyMsg).String(); msg {
		case "ctrl+c":
			return g, tea.Quit
		case "up", "w":
			if msg == "w" {
				player = g.P2
			}
			if player.x > 0 && g.Board.Arr[player.x-1][player.y] != 1 {
				player.x--
			}
		case "down", "s":
			if msg == "s" {
				player = g.P2
			}
			if player.x < g.Board.Height-1 && g.Board.Arr[player.x+1][player.y] != 1 {
				player.x++
			}
		case "left", "a":
			if msg == "a" {
				player = g.P2
			}
			if player.y > 0 && g.Board.Arr[player.x][player.y-1] != 1 {
				player.y--
			}
		case "right", "d":
			if msg == "d" {
				player = g.P2
			}
			if player.y < g.Board.Width-1 && g.Board.Arr[player.x][player.y+1] != 1 {
				player.y++
			}
		case " ":
			if g.Board.Arr[player.x][player.y] == 0 && !player.freezeBomb {
				g.Board.Arr[player.x][player.y] = 2
				player.freezeBomb = true
				return g, tickUnfreeze(player, player.x, player.y)
			}
		}
	case unfreezeMsg:
		p := msg.(unfreezeMsg).P
		x, y := msg.(unfreezeMsg).bombX, msg.(unfreezeMsg).bombY
		p.freezeBomb = false
		return g, tickBomb(x, y)
	case bombMsg:
		x, y := msg.(bombMsg).x, msg.(bombMsg).y
		g.Board.Arr[x][y] = 0
		explosion := fillExplode(g.Board, x, y)
		return g, tickExplode(explosion)
	case exploseMsg:
		points := msg.(exploseMsg).e
		fillLand(g.Board, points)
	}

	return g, nil
}

func tickBomb(x, y int) tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return bombMsg{x, y}
	})
}

func tickUnfreeze(p *player, bombX, bombY int) tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return unfreezeMsg{p, bombX, bombY}
	})
}

func tickExplode(e explosion) tea.Cmd {
	return tea.Tick(_explode_t, func(t time.Time) tea.Msg {
		return exploseMsg{e}
	})
}

func fillExplode(board board, x, y int) explosion {
	points := make([][]int, 0)
	for up := x - 1; up >= 0 && board.Arr[up][y] < 1; up-- {
		board.Arr[up][y]--
		points = append(points, []int{up, y})
	}
	for down := x + 1; down < board.Height && board.Arr[down][y] < 1; down++ {
		board.Arr[down][y]--
		points = append(points, []int{down, y})
	}
	for left := y - 1; left >= 0 && board.Arr[x][left] < 1; left-- {
		board.Arr[x][left]--
		points = append(points, []int{x, left})
	}
	for right := y + 1; right < board.Width && board.Arr[x][right] < 1; right++ {
		board.Arr[x][right]--
		points = append(points, []int{x, right})
	}

	board.Arr[x][y] = -1
	return append(points, []int{x, y})
}

func fillLand(b board, points explosion) {
	for i := 0; i < len(points); i++ {
		b.Arr[points[i][0]][points[i][1]]++
	}
}

func main() {
	p := tea.NewProgram(
		initGame(os.Args[1]),
	)
	if err := p.Start(); err != nil {
		panic(err)
	}
}
