package fpath

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	Separator     = os.PathSeparator
	ListSeparator = os.PathListSeparator
)

type Path string

// New creates a new path
func New(path string)  *Path{
	p := Path(path)
	return &p
}

// Join returns a new path by joining multiple Path/string elements
func Join(elem ...string) *Path {
	return New(path.Join(elem...))
}

// Expand returns a new path with expanded Environment
func Expand(path string)  *Path{
	return New(os.ExpandEnv(path))
}

// Cwd returns current working directory
func Cwd() *Path{
	wd, err := os.Getwd()
	if err != nil {
		return nil
	}
	return New(wd)
}

// FromUrl returns a path with contents initiated from url. Default target is basename of url.
func FromUrl(url string, target ...string) *Path{
	if len(target) < 1{
		target = append(target, filepath.Base(url))
	}
	p := Expand(target[0])
	p.DownloadFrom(url)
	return p
}

// String returns path as string
func (p *Path) String() string {
	return string(*p)
}

// Join returns a new path by joining multiple Path/string elements
func (p *Path) Join(elem ...string)  *Path{
	return Join(p.String(), path.Join(elem...))
}

// Expand returns a new path with expanded Environment
func (p *Path) Expand()  *Path{
	return New(os.ExpandEnv(p.String()))
}

// Abs returns the absolute path
func (p *Path) Abs() *Path{
	path, _ := filepath.Abs(p.String())
	return New(path)
}

// Parent returns the parent path
func (p *Path) Parent() *Path{
	return New(filepath.Dir(p.String()))
}

// Parents returns nth level parent of the path
func (p *Path) Parents(level int) *Path{
	path := p.String()
	for ; level >= 0; level-- {
		path = filepath.Dir(path)
	}
	return New(path)
}

// Base returns the basename of path
func (p *Path) Base() string{
	return filepath.Base(p.String())
}

// Base returns the stem of basename
func (p *Path) Stem() string{
	base := p.Base()
	if i := strings.LastIndex(base, "."); i > 0 {
		return base[:i]
	}
	return base
}

// Base returns the extension of path
func (p *Path) Ext() string{
	return filepath.Ext(p.Base())
}

// Base returns the directory of path
func (p *Path) Dir() string{
	return filepath.Dir(p.String())
}

// Base returns path with new suffix
func (p *Path) WithSuffix(suffix string) *Path{
	return p.Parent().Join(p.Stem() + suffix)
}

// Base returns path with new prefix
func (p *Path) WithPrefix(prefix string) *Path{
	return p.Parent().Join(prefix + p.Base())
}

// Stat returns the FileInfo of the path
func (p *Path) Stat() os.FileInfo{
	s, err := os.Stat(p.String())
	if err != nil {
		return nil
	}
	return s
}

// Size returns the file size of the path
func (p *Path) Size() int64{
	if s, err := os.Stat(p.String()); err == nil {
		s.Size()
	}
	return -1
}

// PrettySize returns the file size of the path
func (p *Path) PrettySize() string{
	if s, err := os.Stat(p.String()); err == nil {
		return PrettySize(s.Size())
	}
	return "0 B"
}
// Exists checks path is realized
func (p *Path) Exists() bool{
	_, err := os.Stat(p.String())
	return err == nil || os.IsExist(err)
}

// IsDir checks if path is a directory
func (p *Path) IsDir() bool{
	s := p.Stat()
	return s != nil && s.IsDir()
}

// IsFile checks if path is a file
func (p *Path) IsFile() bool{
	s := p.Stat()
	return s != nil && !s.IsDir()
}

// ReadLink reads the a symlink
func (p *Path) ReadLink() *Path  {
	src, err := os.Readlink(p.String())
	if err != nil {
		return nil
	}
	return New(src)
}

// Touch trys to create the path as a file
func (p *Path) Touch() error {
	f, err := os.Create(p.String())
	f.Close()
	return err
}

// Remove removes the path. If path is a directory, it must be empty
func (p *Path) Remove() error {
	err := os.Remove(p.String())
	return err
}

// RemoveAll removes the path and all children
func (p *Path) RemoveAll() error {
	err := os.RemoveAll(p.String())
	return err
}

// MkDir crates a directory optionally with its parents
func (p *Path) MkDir(perm os.FileMode, parents bool) error {
	if parents {
		return os.MkdirAll(p.String(), perm)
	} else {
		return os.Mkdir(p.String(), perm)
	}
}

// ReadDir returns all files in the path
func (p *Path) ReadDir() []os.DirEntry  {
	files, err := os.ReadDir(p.String())
	if err != nil {
		return nil
	}
	return files
}

// ListDir returns all files in the path as Paths optionally filtering the hidden
func (p *Path) ListDir(hidden bool) []Path  {
	files := p.ReadDir()
	dir := p.String()
	paths := make([]Path, 0, len(files))
	for _, f := range files{
		if !hidden && strings.Index(f.Name(), ".") == 0 {
			continue
		}
		paths = append(paths, *Join(dir, f.Name()))
	}
	return paths
}

// Glob returns all files matching pattern in the path as Paths
func (p *Path) Glob(pattern string) []Path  {
	files, err := filepath.Glob(p.Join(pattern).String())
	if err != nil {
		return nil
	}
	paths := make([]Path, 0, len(files))
	for _, f := range files{
		paths = append(paths, *New(f))
	}
	return paths
}

// Find all files matching regex pattern in the path and invokes handler
func (p *Path) Find(regex string, handler func(Path)) bool {
	rxp, err := regexp.Compile(regex)
	if err != nil{
		return false
	}
	files := p.ReadDir()
	dir := p.String()
	for _, f := range files{
		fn := f.Name()
		if  rxp.MatchString(fn){
			handler(*Join(dir, fn))
		}
	}
	return len(files) > 0
}

func (p *Path) Match(pattern string) bool  {
	match, err := path.Match(pattern, p.String())
	return err == nil && match
}

func (p *Path) DownloadFrom(url string) error {
	if p.Exists() {
		return nil
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(p.String())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}



type OpenFlag uint32

const (
	ForReading = OpenFlag(os.O_RDONLY)
	ForWriting = OpenFlag(os.O_WRONLY | os.O_CREATE)
	ForReadWrite = OpenFlag(os.O_RDWR | os.O_CREATE)
	ForAppending = OpenFlag(os.O_RDWR | os.O_CREATE | os.O_APPEND)
	ForNewWrite = OpenFlag(os.O_WRONLY | os.O_CREATE | os.O_TRUNC)
)


// ReadBytes reads the contents of file as bytes
func (p *Path) Open(mode OpenFlag) (*os.File, error)  {
	return os.OpenFile(p.String(), int(mode), 0644)
}

// ReadBytes reads the contents of file as bytes
func (p *Path) ReadBytes() []byte  {
	if content, err := os.ReadFile(p.String()); err == nil{
		return content
	}
	return nil
}

// WriteBytes writes the contents to a file as bytes
func (p *Path) WriteBytes(content []byte) error  {
	return os.WriteFile(p.String(), content, 0644)
}

// ReadText reads the contents of file as string
func (p *Path) ReadText() string  {
	if content, err := os.ReadFile(p.String()); err == nil{
		return string(content)
	}
	return ""
}

// WriteBytes writes the contents to a file as bytes
func (p *Path) WriteText(content string) error  {
	return os.WriteFile(p.String(), []byte(content), 0644)
}

// ReadJSON reads the JSON content fom file
func (p *Path) ReadJson() interface{} {
	content := new(interface{})
	json.Unmarshal(p.ReadBytes(), content)
	return *interface{}(content).(*interface{})
}

// ReadJsonMap reads the JSON object from file
func (p *Path) ReadJsonMap() *map[string]interface{} {
	content := new(map[string]interface{})
	json.Unmarshal(p.ReadBytes(), content)
	return content
}

// WriteJSON writes the contents to a file as JSON
func (p *Path) WriteJson(content interface{}) error  {
	jsonBytes, err := json.Marshal(content);
	if err != nil {
		return err
	}
	return os.WriteFile(p.String(), jsonBytes, 0644)
}

func (p *Path) ReadKV(sep string) ValueMap  {
	if fd, err := p.Open(ForReading); err == nil{
		return LoadValueMap(fd, sep, false, false, false)
	}
	return ValueMap{}
}


type Value string

func (v *Value) String() string {
	return string(*v)
}

func (v *Value) Int() int {
	val, _ := strconv.Atoi(v.String());
	return val
}

func (v *Value) Float() float32 {
	val, _ := strconv.ParseFloat(v.String(), 32);
	return float32(val)
}

func (v *Value) Bool() bool {
	val, _ := strconv.ParseBool(v.String());
	return val
}

func (v *Value) Path() *Path {
	p := New(v.String())
	return p
}

func (v *Value) Array(sep string) []Value {
	s := v.String()
	values := []Value{}
	i := strings.Index(s, sep);
	for {
		if i == -1  && len(s) > 0{
			values = append(values, Value(s))
			break
		}
		values = append(values, Value(s[:i]))
		s = s[i + len(sep):]
		i = strings.Index(s, sep);
	}
	return values
}

type ValueMap map[string] *Value

func LoadValueMap(fd io.Reader, sep string, unquote, expandVars, setEnv bool)  ValueMap{
	kvmap := make(ValueMap)
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty line or comments (starting with # or //)
		if len(line) < 3 || line[0] == '#' || line[0:2] == "//" {
			continue
		}

		// Skip lines without separator or no values
		isep := strings.Index(line, sep)
		if isep < 0 || len(line) < isep + 1{
			continue
		}

		//Read key & value
		key := strings.TrimSpace(line[:isep])
		val := strings.TrimSpace(line[isep+1:])
		if unquote {
			if str, err := strconv.Unquote(val); err == nil {
				val = str
			}
		}

		// Expand variables either from source or environment
		if expandVars {
			val = os.Expand(val, func(k string) string {
				if ev, ok := kvmap[k]; ok {
					return ev.String()
				}
				return os.Getenv(k)
			})
		}

		// Export vars to environment
		if setEnv {
			os.Setenv(key, val)
		}

		//fmt.Println(key, "=", val)
		v := Value(val)
		kvmap[key] = &v
	}
	return kvmap
}


// PrettySize formats size to IEC units
func PrettySize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}



