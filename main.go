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
	_bomb   = "☸"
	_player = "⚘"
	_dead   = "☠"
	_wall   = "☲"

	timeout = time.Second * 5
)

type bombMsg struct {
	x int
	y int
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
	x int
	y int
}

type bomb struct {
	x int
	y int
}

type explosion [][]int

type game struct {
	Player player
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
	return game{
		Player: player{1, 1},
		Board:  initBoard(filePath),
	}
}

func (g game) Init() tea.Cmd { return nil }

func (g game) View() string {
	boardBuilder := make([]string, g.Board.Height)
	for i := 0; i < g.Board.Height; i++ {
		rowBuilder := make([]string, g.Board.Width)
		for j := 0; j < g.Board.Width; j++ {
			if i == g.Player.x && j == g.Player.y {
				rowBuilder[j] = _player
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
	switch msg.(type) {
	case tea.KeyMsg:
		switch msg.(tea.KeyMsg).String() {
		case "ctrl+c":
			return g, tea.Quit
		case "up":
			if g.Player.x > 0 && g.Board.Arr[g.Player.x-1][g.Player.y] != 1 {
				g.Player.x--
			}
		case "down":
			if g.Player.x < g.Board.Height-1 && g.Board.Arr[g.Player.x+1][g.Player.y] != 1 {
				g.Player.x++
			}
		case "left":
			if g.Player.y > 0 && g.Board.Arr[g.Player.x][g.Player.y-1] != 1 {
				g.Player.y--
			}
		case "right":
			if g.Player.y < g.Board.Width-1 && g.Board.Arr[g.Player.x][g.Player.y+1] != 1 {
				g.Player.y++
			}
		case " ":
			if g.Board.Arr[g.Player.x][g.Player.y] == 0 {
				g.Board.Arr[g.Player.x][g.Player.y] = 2
				return g, tickBomb(g.Player.x, g.Player.y)
			}
		}
	case bombMsg:
		bomb := msg.(bombMsg)
		g.Board.Arr[bomb.x][bomb.y] = 0
		explosion := fillExplode(g.Board, bomb.x, bomb.y)
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

func tickExplode(e explosion) tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
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

	board.Arr[x][y]--
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
