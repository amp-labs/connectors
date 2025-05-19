package core

type Command string

func FormatCommand(serviceName string, command string) Command {
	return Command(serviceName + "." + command)
}

func (c Command) String() string {
	return string(c)
}
