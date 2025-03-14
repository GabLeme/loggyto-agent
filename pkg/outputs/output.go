package outputs

// Output define a interface para saídas de log
type Output interface {
	Write(logEntry string)
	Close()
}
