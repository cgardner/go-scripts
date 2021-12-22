package main

import (
	"fmt"
	"os"

	"github.com/naggie/dstask"
)

type App struct {
	Config  dstask.Config
	Query   dstask.Query
	Context dstask.Query
	TS      *dstask.TaskSet
}

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
	switch args[0] {
	case "interview":
		template, err := a.findTemplateBySummary("Interview")
		if err != nil {
			dstask.ExitFail(err.Error())
		}

		task := dstask.Task{
			WritePending: true,
			Status:       dstask.STATUS_PENDING,
			Summary:      args[1],
			Tags:         template.Tags,
			Project:      template.Project,
			Priority:     template.Priority,
			Notes:        template.Notes,
		}
		a.AddTask(task)

	case "projects":
		MustNotFail(dstask.CommandShowProjects(a.Config, a.Context, a.Query))
	case "templates":
		MustNotFail(dstask.CommandShowTemplates(a.Config, a.Context, a.Query))
	case "active":
		MustNotFail(dstask.CommandShowActive(a.Config, a.Context, a.Query))
	case "paused":
		MustNotFail(dstask.CommandShowPaused(a.Config, a.Context, a.Query))
	case "today":
		a.Query.Tags = []string{"today"}
		a.Next()
	case "list", "ls":
		fallthrough
	default:
		a.Query = dstask.ParseQuery(args...)
		a.Next()
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
	MustNotFail(dstask.CommandNext(a.Config, a.Context, a.Query))
}

func (a *App) AddTask(t dstask.Task) {
	ts := a.TS
  t = ts.LoadTask(t)
	ts.SavePendingChanges()
	dstask.MustGitCommit(a.Config.Repo, "Added %s", t)
}
