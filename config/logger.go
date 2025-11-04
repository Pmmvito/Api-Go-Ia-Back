package config

import (
	"io"
	"log"
	"os"
)

// Logger fornece uma interface de log estruturada com diferentes níveis (Debug, Info, Warning, Error).
type Logger struct {
	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
	err     *log.Logger
	writer  io.Writer
}

// NewLogger cria e retorna uma nova instância de Logger.
// Ele recebe uma string de prefixo 'p', que é tipicamente o nome do pacote,
// para identificar a origem das mensagens de log.
func NewLogger(p string) *Logger {
	writer := io.Writer(os.Stdout)
	logger := log.New(writer, p, log.Ldate|log.Ltime)

	return &Logger{
		debug:   log.New(writer, "DEBUG: ", logger.Flags()),
		info:    log.New(writer, "INFO: ", logger.Flags()),
		warning: log.New(writer, "WARNING: ", logger.Flags()),
		err:     log.New(writer, "ERROR: ", logger.Flags()),
		writer:  writer,
	}
}

// Debug registra uma mensagem no nível DEBUG.
// Os argumentos são tratados da mesma forma que em fmt.Println.
func (l *Logger) Debug(v ...interface{}) {
	l.debug.Println(v...)
}

// Info registra uma mensagem no nível INFO.
// Os argumentos são tratados da mesma forma que em fmt.Println.
func (l *Logger) Info(v ...interface{}) {
	l.info.Println(v...)
}

// Warn registra uma mensagem no nível WARNING.
// Os argumentos são tratados da mesma forma que em fmt.Println.
func (l *Logger) Warn(v ...interface{}) {
	l.warning.Println(v...)
}

// Error registra uma mensagem no nível ERROR.
// Os argumentos são tratados da mesma forma que em fmt.Println.
func (l *Logger) Error(v ...interface{}) {
	l.err.Println(v...)
}

// Debugf registra uma mensagem formatada no nível DEBUG.
// Os argumentos são tratados da mesma forma que em fmt.Printf.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.debug.Printf(format, v...)
}

// InfoF registra uma mensagem formatada no nível INFO.
// Os argumentos são tratados da mesma forma que em fmt.Printf.
func (l *Logger) InfoF(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

// WarnF registra uma mensagem formatada no nível WARNING.
// Os argumentos são tratados da mesma forma que em fmt.Printf.
func (l *Logger) WarnF(format string, v ...interface{}) {
	l.warning.Printf(format, v...)
}

// ErrorF registra uma mensagem formatada no nível ERROR.
// Os argumentos são tratados da mesma forma que em fmt.Printf.
func (l *Logger) ErrorF(format string, v ...interface{}) {
	l.err.Printf(format, v...)
}
