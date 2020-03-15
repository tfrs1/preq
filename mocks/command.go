package mocks

import "github.com/spf13/pflag"

type Command struct {
	FlagsValue *pflag.FlagSet
}

func (cmd *Command) Flags() *pflag.FlagSet { return cmd.FlagsValue }
