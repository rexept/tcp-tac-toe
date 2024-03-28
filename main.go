package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

type Player struct {
	Index      int
	Connection net.Conn
	Symbol     string
	Score      int
}

// will only handle 2 players for now
var players = []Player{}
var currentPlayer Player

var colors = map[string][2]string{
	"orange": {"\x1b[34m", "\x1b[0m"},
	"cyan":   {"\x1b[36m", "\x1b[0m"},
}

func colorize(text string, color string) string {
	return fmt.Sprintf("%s%s%s", colors[color][0], text, colors[color][1])
}

func closeConnection(c net.Conn) {
	for i, p := range players {
		if p.Connection == c {
			players = append(players[:i], players[i+1:]...)
			break
		}
	}
	c.Close()
}

func handleConnection(conn net.Conn, b map[int]string) {
	scanner := bufio.NewScanner(conn)
	for {
		ok := scanner.Scan()
		if !ok {
			break
		}
		play(scanner.Text(), b, conn)
	}
	closeConnection(conn)

}

func switchCurrentPlayer() {
	if currentPlayer.Index == 0 {
		currentPlayer = players[1]
	} else {
		currentPlayer = players[0]
	}
}

func dispatchBoard(b map[int]string) {
	for _, p := range players {
		output := printBoard(b)
		p.Connection.Write([]byte(output + "\n"))
		if p.Connection == currentPlayer.Connection {
			p.Connection.Write([]byte("your turn "))
		}
	}
}

func play(pos string, b map[int]string, c net.Conn) {
	fmt.Println("> " + pos)
	position, _ := strconv.Atoi(pos)

	if c == currentPlayer.Connection {
		b[position] = currentPlayer.Symbol
		switchCurrentPlayer()
	}

	dispatchBoard(b)
}

func initBoard() map[int]string {
	board := make(map[int]string)
	for i := 0; i < 9; i++ {
		board[i] = fmt.Sprintf("%d", i)
	}
	return board
}

func printBoard(b map[int]string) string {
	output := ""
	for i := 0; i < 9; i += 3 {
		output += fmt.Sprintf("%v | %v | %v\n", b[i], b[i+1], b[i+2])
	}
	return output
}

func main() {
	listener, _ := net.Listen("tcp", "localhost:8080")
	fmt.Println("Listening on localhost:8080.")
	b := initBoard()
	defer listener.Close()

	for {
		conn, _ := listener.Accept() // connect using telnet cmd: telnet localhost 8080
		fmt.Printf("client connected from %v\n", conn.RemoteAddr().String())

		if len(players) < 2 {
			player := Player{
				Index:      len(players),
				Connection: conn,
				Symbol:     []string{colorize("x", "orange"), colorize("o", "cyan")}[len(players)],
				Score:      0,
			}
			players = append(players, player)
			currentPlayer = player
			go handleConnection(conn, b)
		}
	}
}
