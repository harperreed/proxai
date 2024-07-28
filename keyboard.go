package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func handleKeyboardInput(s *ProxyServer) {
    reader := bufio.NewReader(os.Stdin)
    for {
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        switch input {
        case "r", "reset":
            s.resetCounters()
            fmt.Println("Counters reset.")
        case "c", "clear":
            clearConsole()
            fmt.Println("Console cleared.")
        case "q", "quit":
            fmt.Println("Exiting...")
            quit <- os.Interrupt
            return
        default:
            fmt.Println("Unknown command. Available commands: r/reset, c/clear, q/quit")
        }
    }
}

func clearConsole() {
    fmt.Print("\033[2J")
    fmt.Print("\033[H")
}
