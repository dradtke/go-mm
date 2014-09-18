package main

import (
	commander "code.google.com/p/go-commander"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	DEBUG = false
)

var (
	MM_ROOT string

	DEFAULT_PACKAGE = map[string]string{
		"ApexClass":      "*",
		"ApexComponent":  "*",
		"ApexPage":       "*",
		"StaticResource": "*",
		"ApexTrigger":    "*",
	}

	NEW_PROJECT = commander.Command{
		Run: func(cmd *commander.Command, args []string) {
			opts := new(struct {
				ProjectName string            `json:"project_name"`
				Username    string            `json:"username"`
				Password    string            `json:"password"`
				OrgType     string            `json:"org_type"`
				Package     map[string]string `json:"package"`
			})

			flagset := flag.NewFlagSet("new_project", flag.ContinueOnError)
			flagset.StringVar(&opts.ProjectName, "project_name", "", "")
			flagset.StringVar(&opts.Username, "username", "", "")
			flagset.StringVar(&opts.Password, "password", "", "")
			flagset.StringVar(&opts.OrgType, "org_type", "", "")
			flagset.Parse(args)

			opts.Package = DEFAULT_PACKAGE
			if opts.OrgType == "" {
				opts.OrgType = "sandbox"
			}
			var res GenericResult
			ExecMM("new_project", &opts, &res)
			if !res.Success {
				fmt.Fprintln(os.Stderr, "New project creation failed.")
				os.Exit(1)
			} else {
				fmt.Println("Success.")
				os.Exit(0)
			}
		},

		UsageLine:   "new_project",
		Short:       "nothing here!",
		Long:        "nothing here either!",
		CustomFlags: true,
	}

	COMPILE = commander.Command{
		Run: func(cmd *commander.Command, args []string) {
			opts := new(struct {
				ProjectName string   `json:"project_name"`
				Files       []string `json:"files"`
			})

			flagset := flag.NewFlagSet("new_project", flag.ContinueOnError)
			flagset.StringVar(&opts.ProjectName, "project_name", "", "")
			flagset.Parse(args)

			opts.Files = make([]string, flagset.NArg())
			for i, arg := range args[len(args)-flagset.NArg():] {
				if !filepath.IsAbs(arg) {
					var err error
					arg, err = filepath.Abs(arg)
					if err != nil {
						panic(err)
					}
				}
				opts.Files[i] = arg
			}

			var res CompileResult
			ExecMM("compile", &opts, &res)
			if !res.Success {
				fmt.Fprintf(os.Stderr, "Line %s: %s\n",
					res.Details.ComponentFailures.LineNumber,
					res.Details.ComponentFailures.Problem)
				os.Exit(1)
			} else {
				fmt.Println("Success.")
				os.Exit(0)
			}
		},

		UsageLine:   "compile",
		Short:       "nothing here!",
		Long:        "nothing here either!",
		CustomFlags: true,
	}

	TEST = commander.Command{
		Run: func(cmd *commander.Command, args []string) {
			opts := new(struct {
				ProjectName string   `json:"project_name"`
				Classes     []string `json:"classes"`
				RunAll      bool     `json:"run_all_tests"`
			})

			flagset := flag.NewFlagSet("unit_test", flag.ContinueOnError)
			flagset.StringVar(&opts.ProjectName, "project_name", "", "")
			flagset.BoolVar(&opts.RunAll, "run_all_tests", false, "")
			flagset.Parse(args)
			opts.Classes = flagset.Args()

			var res TestResult
			ExecMM("unit_test", &opts, &res)
			fmt.Println(res.Status)
			os.Exit(0)
		},

		UsageLine:   "unit_test",
		Short:       "nothing here!",
		Long:        "nothing here either!",
		CustomFlags: true,
	}
)

type GenericResult struct {
	Success bool
}

type CompileResult struct {
	GenericResult

	Body    string // on success
	Details ResultDetails
}

type ResultDetails struct {
	ComponentFailures ComponentResult
	// TODO: runTestResult
}

type ComponentResult struct {
	FullName   string
	Problem    string
	LineNumber string
}

type TestResult struct {
	Status         string
	ExtendedStatus string
}

func ExecMM(cmd string, input interface{}, output interface{}) {
	proc := exec.Command("python", filepath.Join(MM_ROOT, "mm.py"), cmd)
	stdin, err := proc.StdinPipe()
	if err != nil {
		panic(err)
	}
	stdout, err := proc.StdoutPipe()
	if err != nil {
		panic(err)
	}

	/*
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(input)
		fmt.Printf("%s\n\n", buf.String())
		io.Copy(stdin, &buf)
	*/

	json.NewEncoder(stdin).Encode(input)
	proc.Start()
	stdin.Close()

	if DEBUG {
		data, err := ioutil.ReadAll(stdout)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", string(data))
		os.Exit(0)
	}

	if err := json.NewDecoder(stdout).Decode(&output); err != nil {
		panic(err)
	}

	if err := proc.Wait(); err != nil {
		panic(err)
	}
}

func main() {
	if MM_ROOT = os.Getenv("MM_ROOT"); MM_ROOT == "" {
		fmt.Println("Please set MM_ROOT and try again.")
		os.Exit(1)
	}

	c := commander.Commander{
		Name: "mm",
		Commands: []*commander.Command{
			&NEW_PROJECT,
			&COMPILE,
			&TEST,
		},
		Flag: flag.NewFlagSet("mm", flag.ContinueOnError),
	}

	err := c.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
