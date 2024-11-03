package log
import (
	"fmt"
)

func Error( msg string ) {
	errorMsg := "\x1b[31;1m[ERROR]"
	errorMsg += "\033[0m"

	fmt.Println( errorMsg, msg )
}
