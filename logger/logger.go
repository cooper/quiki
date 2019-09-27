package logger

// L logs some stuff.
func L(s string, stuff ...interface{}) {

}

// Lindent logs some stuff and then increases the indentation level.
func Lindent(s string, stuff ...interface{}) {
	L(s, stuff...)
	Indent()
}

// Indent increases the indentation level.
func Indent() {

}

// Back decreases the indentation level.
func Back() {

}
