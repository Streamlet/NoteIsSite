package note

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/Streamlet/NoteIsSite/util"
	"github.com/fsnotify/fsnotify"
)

type watcherHandler interface {
	FileCreated(path string)
	FileRemoved(path string)
	FileChanged(path string)
}

type watcher struct {
	inner   *fsnotify.Watcher
	dirs    map[string]bool
	closing chan bool
	closed  chan bool
}

func newWatcher() (*watcher, error) {
	innerWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := new(watcher)
	w.inner = innerWatcher
	w.dirs = make(map[string]bool)
	w.closing = make(chan bool, 1)
	w.closed = make(chan bool, 1)
	return w, nil
}

func (w watcher) addDirs(dir string) error {
	err := w.inner.Add(dir)
	if err != nil {
		return err
	}
	w.dirs[dir] = true
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			subPath := dir + "/" + f.Name()
			err = w.addDirs(subPath)
			if err != nil {
				return err
			}
			w.dirs[subPath] = true
		}
	}
	return nil
}

func (w watcher) watch(handler watcherHandler) {
	util.Assert(handler != nil, "handler MUST NOT be nil")
	go func() {
		for {
			select {
			case ev := <-w.inner.Events:
				if ev.Op&fsnotify.Create != 0 {
					if f, err := os.Stat(ev.Name); err == nil {
						if f.IsDir() {
							_ = w.addDirs(ev.Name)
						}
						go handler.FileCreated(ev.Name)
					}
				}
				if ev.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
					_ = w.inner.Remove(ev.Name)
					go handler.FileRemoved(ev.Name)
				}
				if ev.Op&fsnotify.Write != 0 {
					if f, err := os.Stat(ev.Name); err == nil {
						if !f.IsDir() {
							go handler.FileChanged(ev.Name)
						}
					}
				}
			case err := <-w.inner.Errors:
				log.Println(err.Error())
			case <-w.closing:
				w.closed <- true
				return
			}
		}
	}()
}

func (w watcher) close() error {
	w.closing <- true
	<-w.closed
	return w.inner.Close()
}
