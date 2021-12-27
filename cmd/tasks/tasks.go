package main

import (
	"fmt"
	"os"
	"time"

	"github.com/naggie/dstask"
)

type App struct {
	Config  dstask.Config
	Query   dstask.Query
	Context dstask.Query
	TS      *dstask.TaskSet
}

type CommandFunction func(dstask.Config, dstask.Query, dstask.Query) error

func main() {
	app := NewApp()
	app.Execute(os.Args[1:])
}

func MustNotFail(err error) {
	if err != nil {
		dstask.ExitFail(err.Error())
	}
}

func NewApp() App {
	config := dstask.NewConfig()
	dstask.EnsureRepoExists(config.Repo)

	context := dstask.Query{}
	query := dstask.Query{
		IgnoreContext: true,
	}

	taskSet, err := dstask.LoadTaskSet(config.Repo, config.IDsFile, false)
	MustNotFail(err)

	return App{
		Config:  config,
		Context: context,
		Query:   query,
		TS:      taskSet,
	}
}

func (a *App) Execute(args []string) {
	if len(args) == 0 {
		a.Next()
		return
	}

	switch args[0] {
	case "interview":
		task := a.createTaskFromTemplate(args[1], "Interview")
		a.AddTask(task)
	case "inbox", "in":
		a.TS.FilterOrganised()
		a.TS.DisplayByNext(a.Context, true)
	case "new-hire":
		task := a.createTaskFromTemplate(fmt.Sprintf("New Hire - %s", args[1]), "New Hire")
		a.AddTask(task)
	case "projects":
		MustNotFail(dstask.CommandShowProjects(a.Config, a.Context, a.Query))
	case "templates":
		MustNotFail(dstask.CommandShowTemplates(a.Config, a.Context, a.Query))
	case "active":
		MustNotFail(dstask.CommandShowActive(a.Config, a.Context, a.Query))
	case "paused":
		MustNotFail(dstask.CommandShowPaused(a.Config, a.Context, a.Query))
	case "lift":
		summary := fmt.Sprintf("%s (%s lbs)", args[1], args[2])
		task := a.createTaskFromTemplate(summary, "Lift")
		task.Status = dstask.STATUS_RESOLVED
		task.Resolved = time.Now()
		fmt.Println(task)
		a.AddTask(task)
	case "today":
		a.Query.Tags = []string{"today"}
		a.Next()
	case "list", "ls":
		fallthrough
	default:
		a.DefaultCommand(args)
	}
}

func (a *App) findTemplateBySummary(summary string) (dstask.Task, error) {
	a.TS.UnHide()
	a.TS.FilterByStatus(dstask.STATUS_TEMPLATE)
	tasks := a.TS.Tasks()

	for _, task := range tasks {
		if task.Summary == summary {
			return task, nil
		}
	}

	return dstask.Task{}, fmt.Errorf("Unable to find a matching template")
}

func (a *App) Next() {
	state := dstask.LoadState(a.Config.StateFile)
	MustNotFail(dstask.CommandNext(a.Config, state.Context, a.Query))
}

func (a *App) AddTask(t dstask.Task) {
	ts := a.TS
	t = ts.LoadTask(t)
	ts.SavePendingChanges()
	dstask.MustGitCommit(a.Config.Repo, "Added %s", t)
}

func (a *App) createTaskFromTemplate(summary string, templateSummary string) dstask.Task {
	template, err := a.findTemplateBySummary(templateSummary)
	if err != nil {
		dstask.ExitFail(err.Error())
	}

	task := dstask.Task{
		WritePending: true,
		Status:       dstask.STATUS_PENDING,
		Summary:      summary,
		Tags:         template.Tags,
		Project:      template.Project,
		Priority:     template.Priority,
		Notes:        template.Notes,
	}
	return task
}

func (a *App) DefaultCommand(args []string) {
	/***
	  Tasks Not Covered Here:
	    - Context
	    - Undo
	    - Sync
	    - Git
	    - Versions
	    - Completions
	*/
	commandMap := map[string]CommandFunction{
		dstask.CMD_NEXT:             dstask.CommandNext,
		dstask.CMD_SHOW_OPEN:        dstask.CommandShowOpen,
		dstask.CMD_ADD:              dstask.CommandAdd,
		dstask.CMD_RM:               dstask.CommandRemove,
		dstask.CMD_REMOVE:           dstask.CommandRemove,
		dstask.CMD_TEMPLATE:         dstask.CommandTemplate,
		dstask.CMD_LOG:              dstask.CommandLog,
		dstask.CMD_START:            dstask.CommandStart,
		dstask.CMD_STOP:             dstask.CommandStop,
		dstask.CMD_DONE:             dstask.CommandDone,
		dstask.CMD_RESOLVE:          dstask.CommandDone,
		dstask.CMD_MODIFY:           dstask.CommandModify,
		dstask.CMD_EDIT:             dstask.CommandEdit,
		dstask.CMD_NOTE:             dstask.CommandNote,
		dstask.CMD_NOTES:            dstask.CommandNote,
		dstask.CMD_SHOW_ACTIVE:      dstask.CommandShowActive,
		dstask.CMD_SHOW_PAUSED:      dstask.CommandShowPaused,
		dstask.CMD_OPEN:             dstask.CommandOpen,
		dstask.CMD_SHOW_PROJECTS:    dstask.CommandShowProjects,
		dstask.CMD_SHOW_TAGS:        dstask.CommandShowTags,
		dstask.CMD_SHOW_TEMPLATES:   dstask.CommandShowTemplates,
		dstask.CMD_SHOW_RESOLVED:    dstask.CommandShowResolved,
		dstask.CMD_SHOW_UNORGANISED: dstask.CommandShowUnorganised,
	}

	a.Query = dstask.ParseQuery(args...)
	state := dstask.LoadState(a.Config.StateFile)
	query := dstask.ParseQuery(args...)

	commandFunction, exists := commandMap[query.Cmd]

	if query.Cmd == "" || !exists {
		MustNotFail(dstask.CommandNext(a.Config, state.Context, query))
	}

	MustNotFail(commandFunction(a.Config, state.Context, query))
}
