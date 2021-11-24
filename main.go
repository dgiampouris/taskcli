package main

import (
	"log"
	"os"

	"github.com/dgiampouris/taskcli/task"
	"github.com/urfave/cli/v2"
)

/*
    This is a CLI tool that can be used to manage your TODOs
    in the terminal. The basic commands that are available
    are the ones below:
	- add	    Add new task to your TODO list
	- delete    Delete a task on your list
	- list	    List all of your incomplete tasks
*/

func main() {
	app := cli.NewApp()
	app.Name = `
 /$$$$$$$$ /$$$$$$   /$$$$$$  /$$   /$$ 
|__  $$__//$$__  $$ /$$__  $$| $$  /$$/
   | $$  | $$  \ $$| $$  \__/| $$ /$$/
   | $$  | $$$$$$$$|  $$$$$$ | $$$$$/ 
   | $$  | $$__  $$ \____  $$| $$  $$
   | $$  | $$  | $$ /$$  \ $$| $$\  $$
   | $$  | $$  | $$|  $$$$$$/| $$ \  $$
   |__/  |__/  |__/ \______/ |__/  \__/` + "\n"
	app.Usage = "------------------------------------\n\nA simple CLI task manager"

	app.Commands = []*cli.Command{
		{
			Name:  "add",
			Usage: "Add a new task to your list",
			Action: func(c *cli.Context) error {
				task.AddTask(c.Args().Get(0))
				task.ListTasks()
				return nil
			},
		},
		{
			Name:  "delete",
			Usage: "Delete a task on your list",
			Action: func(c *cli.Context) error {
				task.DeleteTask(c.Args().Get(0))
				task.ListTasks()
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "List all tasks",
			Action: func(c *cli.Context) error {
				task.ListTasks()
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
