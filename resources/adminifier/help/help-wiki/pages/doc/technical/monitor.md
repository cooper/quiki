# monitor
--
    import "github.com/cooper/quiki/monitor"

Package monitor provides a file monitor that pre-generates wiki pages and images
each time a change is detected on the filesystem.

## Usage

#### func  WatchWiki

```go
func WatchWiki(w *wiki.Wiki)
```
WatchWiki starts a file monitor loop for the provided wiki.
