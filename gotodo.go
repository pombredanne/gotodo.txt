package main

import  (
        "fmt"
        "os"
        "../go-todotxt"
        "github.com/spf13/cobra"
        "os/user"
        "strings"
        "strconv"
        "github.com/rakyll/globalconf"
        "flag"
)

func extendedLoader(filename string) (todotxt.TaskList, error) {
        usr, err := user.Current()
        if err != nil {
                return nil, err
        }

        filename = strings.Replace(filename, "~", usr.HomeDir, -1)
        tasks := todotxt.LoadTaskList(filename)

        return tasks, nil
}

func main() {

        conf, _ := globalconf.New("gotodo")

        var numtasks bool
        var sortby string
        var finished bool
        var prettyformat string
        var filename string

        var flagFilename = flag.String("file", "", "Location of the todo.txt file.")

        var flags = make(map[string]string)

        var cmdConfig = &cobra.Command{
            Use:   "config [key] [value]",
            Short: "Show and sets config values",
            Long:  `Config can be used to see and also set configuration variables.`,
            Run: func(cmd *cobra.Command, args []string) {
                    if len(args) == 0{
                            fmt.Printf("Available config variables:\n\n")
                            for k := range flags {
                                    fmt.Printf("%s\n", k)
                            }
                    }
                    if len(args) == 1 {
                            val, exists := flags[args[0]]
                            if exists {
                                    fmt.Printf("%s\n", val)
                            } else {
                                    // otherwise exit with non-zero status
                                    os.Exit(1)
                            }
                    }
                    if len(args) == 2 {
                            _, exists := flags[args[0]]
                            if exists {
                                    f := &flag.Flag{Name: args[0], Value: args[1]}
                                    conf.Set("", f)
                            } else {
                                    os.Exit(1)
                            }
                    }
            },
        }

        var cmdList = &cobra.Command{
            Use:   "list [keyword]",
            Short: "Lists tasks that contain keyword, if any",
            Long:  `List is the most basic command that is used for listing tasks.
                    You can specify a keyword as well as other options.`,
            Run: func(cmd *cobra.Command, args []string) {
                tasks, err := extendedLoader(filename)
                if err != nil {
                        fmt.Println(err)
                        return
                }

                if numtasks {
                    fmt.Println(tasks.Len())
                } else {
                    tasks.Sort(sortby)

                    var filteredTasks todotxt.TaskList
                    for _, task := range tasks {
                        if (!task.Finished() && !finished) ||
                           (task.Finished() && finished) {
                           filteredTasks = append(filteredTasks, task)
                        }
                    }

                    for _, task := range filteredTasks {
                            task.SetIdPaddingBy(tasks)
                            fmt.Println(task.PrettyPrint(prettyformat))
                    }
                }
            },
        }
        cmdList.Flags().BoolVarP(&numtasks, "num-tasks", "n", false,
                                 "Show the number of tasks")
        cmdList.Flags().BoolVarP(&finished, "finished", "f", false,
                                 "Show finished tasks")
        cmdList.Flags().StringVarP(&sortby, "sort", "s", "prio",
                                   "Sort tasks by parameter (prio|date|len|prio-rev|date-rev|len-rev|id|rand)")
        cmdList.Flags().StringVarP(&prettyformat, "pretty", "", "%i %p %t",
                                   "Pretty print tasks")

        var cmdAdd = &cobra.Command{
            Use:   "add [task]",
            Short: "Adds a task to the todo list.",
            Long:  `Adds a task to the todo list.`,
            Run: func(cmd *cobra.Command, args []string) {
                tasks, err := extendedLoader(filename)
                if err != nil {
                        fmt.Println(err)
                        return
                }

                task := strings.Join(args, " ")
                tasks.Add(task)

                tasks.Save(filename)
            },
        }

        var nofinishdate bool
        var cmdDone = &cobra.Command{
            Use:   "done [taskid]",
            Short: "Marks task as done.",
            Long:  `Marks task as done.`,
            Run: func(cmd *cobra.Command, args []string) {
                tasks, err := extendedLoader(filename)
                if err != nil {
                        fmt.Println(err)
                        return
                }

                if len(args) < 1 {
                        fmt.Println("So what needs to be done?")
                        return
                }

                taskid, err := strconv.Atoi(args[0])
                if err != nil {
                        fmt.Printf("Do you really consider that a number? %v\n", err)
                        return
                }

                err = tasks.Done(taskid, !nofinishdate)
                if err != nil {
                        fmt.Printf("There was an error %v\n", err)
                }

                tasks.Save(filename)
            },
        }
        cmdDone.Flags().BoolVarP(&nofinishdate, "no-finish-date", "D", false,
                                        "Do not mark finished tasks with date.")

        var cmdArchive = &cobra.Command{
            Use:   "archive [taskid]",
            Short: "Archives task.",
            Long:  `Archives task.`,
            Run: func(cmd *cobra.Command, args []string) {
                tasks, err := extendedLoader(filename)
                if err != nil {
                        fmt.Println(err)
                        return
                }

                if len(args) < 1 {
                        fmt.Println("So what needs to be done?")
                        return
                }

                taskid, err := strconv.Atoi(args[0])
                if err != nil {
                        fmt.Printf("Do you really consider that a number? %v\n", err)
                        return
                }

                fmt.Printf("Archiving task %v\n", taskid)

                tasks.Save(filename)
            },
        }


        var editprio string
        var edittodo string
        var cmdEdit = &cobra.Command{
            Use:   "edit [taskid]",
            Short: "Edits given task.",
            Long:  `Edits given task.`,
            Run: func(cmd *cobra.Command, args []string) {
                tasks, err := extendedLoader(filename)
                if err != nil {
                        fmt.Println(err)
                        return
                }

                if len(args) < 1 {
                        fmt.Println("So what do you want to edit?")
                        return
                }

                taskid, err := strconv.Atoi(args[0])
                if err != nil {
                        fmt.Printf("Do you really consider that a number? %v\n", err)
                        return
                }

                if len(editprio) > 0 {
                        tasks[taskid].SetPriority(editprio[0])
                        tasks[taskid].RebuildRawTodo()
                }

                if len(edittodo) > 0 {
                        tasks[taskid].SetTodo(edittodo)
                        tasks[taskid].RebuildRawTodo()
                }

                tasks.Save(filename)
            },
        }

        cmdEdit.PersistentFlags().StringVarP(&editprio, "priority", "p", "",
                                     "Sets task's priority.")
        cmdEdit.PersistentFlags().StringVarP(&edittodo, "todo", "t", "",
                                     "Edit task's todo.")

        var GotodoCmd = &cobra.Command{
            Use:   "gotodo",
            Short: "Gotodo is a go implementation of todo.txt.",
            Long: `A small, fast and fun implementation of todo.txt`,
            Run: func(cmd *cobra.Command, args []string) {
                cmdList.Run(cmd, nil)
            },
        }

        GotodoCmd.PersistentFlags().StringVarP(&filename, "filename", "", "",
                                     "Load tasks from this file.")

        conf.ParseAll()

        // values here
        flags["file"] = *flagFilename

        // sadly, this is the best we can do right now
        if filename == "" {
                if *flagFilename == "" {
                        filename = "todo.txt"
                } else {
                        filename = *flagFilename
                }
        }

        GotodoCmd.AddCommand(cmdList)
        GotodoCmd.AddCommand(cmdAdd)
        GotodoCmd.AddCommand(cmdDone)
        GotodoCmd.AddCommand(cmdArchive)
        GotodoCmd.AddCommand(cmdEdit)
        GotodoCmd.AddCommand(cmdConfig)
        GotodoCmd.Execute()
}
