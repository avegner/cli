package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrUnknownCommand = errors.New("unknown command")
	ErrNotEnoughArgs  = errors.New("not enough arguments")
)

type Element func(cli *CLI)

type command struct {
	usage   string
	argsNum int
	cb      CommandCallback
}

type CommandCallback func(opts OptionMap, args []string) error

func WithCommand(name string, usage string, argsNum int, cb CommandCallback) Element {
	return func(cli *CLI) {
		cli.cmds[name] = command{
			usage,
			argsNum,
			cb,
		}
	}
}

type option struct {
	usage string
	value interface{}
}

func WithBoolOption(name string, usage string, value bool) Element {
	return func(cli *CLI) {
		op := option{
			usage,
			value,
		}
		op.value = flag.Bool(name, value, usage)
		cli.opts[name] = op
	}
}

func GetBoolOption(opts OptionMap, name string) bool {
	return *opts[name].value.(*bool)
}

type CLI struct {
	cmds commandMap
	opts OptionMap
}

// commandMap sets command name - command relation.
type commandMap map[string]command

// OptionMap sets option name - option relation.
type OptionMap map[string]option

func New(elems ...Element) *CLI {
	cli := CLI{
		cmds: make(commandMap),
		opts: make(OptionMap),
	}
	for _, e := range elems {
		e(&cli)
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [option...] command [arg...]\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(os.Stderr, "\nCommand is one of:\n")
		for n, cmd := range cli.cmds {
			fmt.Fprintf(os.Stderr, "%s - %s\n", n, cmd.usage)
		}
		fmt.Fprintln(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
	}

	return &cli
}

func (cli *CLI) Run() error {
	// parse stdin args
	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	name := flag.Args()[0]
	args := flag.Args()[1:]

	cmd, ok := cli.cmds[name]
	if !ok {
		return ErrUnknownCommand
	}

	if cmd.argsNum != 0 && cmd.argsNum > len(args) {
		return ErrNotEnoughArgs
	}

	return cmd.cb(cli.opts, args)
}
