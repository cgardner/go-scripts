package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bitfield/script"
)

type App struct {
	ConfigFile string
}

func main() {
	args := os.Args[1:]

	ledgerFile := ""
	if len(args) >= 1 {
		ledgerFile = args[0]
	}

	action := "help"
	if len(args) >= 2 {
		action = args[2]
	}

	actionArgs := []string{}
	if len(args) >= 3 {
		actionArgs = args[3:]
	}

	app := NewApp(ledgerFile)
	app.Execute(action, actionArgs...)
}

func monthInterval(y int, m time.Month) (firstDay, lastDay time.Time) {
	firstDay = time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	lastDay = time.Date(y, m+1, 1, 0, 0, 0, -1, time.UTC)
	return firstDay, lastDay
}
func yearInterval(y int) (firstDay, lastDay time.Time) {
	firstDay = time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
	lastDay = time.Date(y+1, 1, 1, 0, 0, 0, -1, time.UTC)
	return firstDay, lastDay
}

func NewApp(configFile string) App {
	return App{
		ConfigFile: configFile,
	}
}

func (a *App) Execute(action string, args ...string) {
	if len(action) == 0 {
		return
	}

	command := ""

	y, m, _ := time.Now().Date()
	switch action {
	case "copy":
		entryName := args[0]
		command = a.generateLedgerExecution(fmt.Sprintf("entry %s", entryName))
	case "current", "cur":
		firstDay, lastDay := monthInterval(y, m)
		command = a.generateLedgerExecution(fmt.Sprintf("balance -b %s -e %s", firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02")))
	case "previous", "prev":
		firstDay, lastDay := monthInterval(y, m-1)
		command = a.generateLedgerExecution(fmt.Sprintf("balance -b %s -e %s", firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02")))
	case "bal":
		command = a.generateLedgerExecution("balance -C")
	default:
		defaultArgs := append([]string{action}, args...)
		command = a.generateLedgerExecution(strings.Join(defaultArgs, " "))
	}

	script.Exec(command).Stdout()
}

func (a *App) generateLedgerExecution(args string) string {
	return fmt.Sprintf("ledger --color -f %s %s", a.ConfigFile, args)
}

func ExitFail(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31m"+format+"\033[0m\n", a...)
	os.Exit(1)
}
