package watcher

import (
	"github.com/fsnotify/fsnotify"
	"github.com/naharp/fpath"
	"log"
	"path"
	"time"
)


// Handler function is called when file changes. Return to true to chain to next matching handler
type Handler func(action string, target fpath.Path) bool

type EventMap map[string] Handler

func Watch(e EventMap) *fsnotify.Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	go func() {
		mtimes := map[string] time.Time{}
		for {
			select {
			case ev := <-watcher.Events:
				for pattern, handler := range e {
					match, err := path.Match(pattern, path.Base(ev.Name))
					if match && err == nil {
						fp := fpath.Path(ev.Name)
						stat := fp.Stat()
						lastMtime, exists := mtimes[ev.Name]
						if !exists || stat == nil || lastMtime != stat.ModTime() {
							if stat != nil{
								mtimes[ev.Name] = stat.ModTime()
							}
							if !handler(ev.Op.String(), fp) {
								break
							}
						}
					}
				}
			case err := <-watcher.Errors:
				// Nothing to do with errors
				log.Println(err)
			}
		}
	}()
	return watcher
}

//
//func main()  {
//	p := fpath.Expand("$HOME")
//	p.Dir()
//	Watch(p, EventMap{
//		"*.css": func(action string, file fpath.Path) bool {
//			log.Println(action, file)
//			return true
//		},
//	})
//}
