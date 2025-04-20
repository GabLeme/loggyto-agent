package outputs

type Output interface {
	Write(logEntry string)
	Close()
}
