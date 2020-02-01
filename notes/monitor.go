package notes

import (
	"github.com/Streamlet/NoteIsSite/util"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"os"
)

type watcherHandler interface {
	FileCreated(path string)
	FileRemoved(path string)
	FileChanged(path string)
}

func watchDir(dir string, handler watcherHandler) error {
	w, err := newWatcher()
	if err != nil {
		return err
	}

	err = w.addDirs(dir)
	if err != nil {
		return err
	}

	w.watch(handler)

	return nil
}

type watcher struct {
	inner *fsnotify.Watcher
	dirs  map[string]bool
}

func newWatcher() (*watcher, error) {
	innerWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := new(watcher)
	w.inner = innerWatcher
	w.dirs = make(map[string]bool)
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
						handler.FileCreated(ev.Name)
					}
				}
				if ev.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
					_ = w.inner.Remove(ev.Name)
					handler.FileRemoved(ev.Name)
				}
				if ev.Op&fsnotify.Write != 0 {
					if f, err := os.Stat(ev.Name); err == nil {
						if !f.IsDir() {
							handler.FileChanged(ev.Name)
						}
					}
				}
			case <-w.inner.Errors:
				break
			}
		}
	}()
}
