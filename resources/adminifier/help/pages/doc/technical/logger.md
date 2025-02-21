# logger
--
    import "github.com/cooper/quiki/logger"


## Usage

#### func  Back

```go
func Back()
```
Back decreases the indentation level.

#### func  Indent

```go
func Indent()
```
Indent increases the indentation level.

#### func  L

```go
func L(s string, stuff ...interface{})
```
L logs some stuff.

#### func  Lindent

```go
func Lindent(s string, stuff ...interface{})
```
Lindent logs some stuff and then increases the indentation level.
