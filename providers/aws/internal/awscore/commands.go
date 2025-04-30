package awscore

import "github.com/amp-labs/connectors/internal/datautils"

type Command string

func FormatCommand(serviceName string, registry datautils.Map[string, string], objectName string) Command {
	return Command(serviceName + "." + registry[objectName])
}

func (c Command) String() string {
	return string(c)
}
