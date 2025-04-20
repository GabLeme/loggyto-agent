package outputs

import "fmt"

type StdoutOutput struct{}

func NewStdoutOutput() *StdoutOutput {
	return &StdoutOutput{}
}

func (o *StdoutOutput) Write(logEntry string) {
	fmt.Println(logEntry)
}

func (o *StdoutOutput) Close() {
	fmt.Println("Closing Stdout Output...")
}
