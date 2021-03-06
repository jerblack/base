package base

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func P(s string, i ...any) {
	now := time.Now()
	t := strings.ToLower(strings.TrimRight(now.Format("3.04PM"), "M"))
	notice := fmt.Sprintf("%s | %s", t, fmt.Sprintf(s, i...))
	fmt.Println(notice)
}
func ChkFatal(err error, msgs ...string) {
	if err != nil {
		for _, msg := range msgs {
			fmt.Println(msg)
		}
		fmt.Println("----------------------")
		defer fmt.Println("----------------------")
		panic(err)
	}
}
func Chk(err error, msgs ...string) {
	if err != nil {
		for _, msg := range msgs {
			fmt.Println(msg)
		}
		fmt.Println("----------------------")
		fmt.Println(err)
		fmt.Println("----------------------")
	}
}

func ErrMsg(e error, msgs ...string) error {
	if e != nil {
		errs := []string{e.Error()}
		for _, msg := range msgs {
			errs = append(errs, msg)
		}
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func HumanSize(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%8dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%6.1f%cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func ArrayIdx(slice []string, val string) int {
	for n, item := range slice {
		if item == val {
			return n
		}
	}
	return -1
}

func IsAny[T comparable](a T, b ...T) bool {
	for _, _b := range b {
		if a == _b {
			return true
		}
	}
	return false
}

func IsAnyStr(a string, b ...string) bool {
	for _, _b := range b {
		if a == _b {
			return true
		}
	}
	return false
}
func IsAnyInt(a int, b ...int) bool {
	for _, _b := range b {
		if a == _b {
			return true
		}
	}
	return false
}
func HasString(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}
func HasInt(arr []int, s int) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

func Run(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	return cmd.Run()
}
func PrintCmd(cmd []string) {
	var parts []string
	for _, c := range cmd {
		if strings.Contains(c, " ") {
			c = fmt.Sprintf("\"%s\"", c)
		}
		parts = append(parts, c)
	}
	P(strings.Join(parts, " "))
}
func RmEmptyFolders(root string) {
	var folders []string
	root, _ = filepath.Abs(root)
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			path, _ = filepath.Abs(path)
			if root != path {
				folders = append(folders, path)
			}
		}
		return nil
	}
	err := filepath.Walk(root, walk)
	ChkFatal(err)

	fn := func(i, j int) bool {
		// reverse sort
		return len(folders[j]) < len(folders[i])
	}
	sort.Slice(folders, fn)
	for _, f := range folders {
		if IsDirEmpty(f) {
			err = os.Remove(f)
			Chk(err)
		}
	}
}
func IsDirEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer func() {
		err = f.Close()
		Chk(err)
	}()

	// read in ONLY one file
	_, err = f.Readdir(1)

	// if file is EOF the dir is empty.
	if err == io.EOF {
		return true
	}
	if err == io.EOF {
		return true
	}
	return false
}
func GetLegalFilename(str string) string {
	reControlCharsRegex := regexp.MustCompile("[\u0000-\u001f\u0080-\u009f]")

	// https://github.com/sindresorhus/filename-reserved-regex/blob/master/index.js
	filenameReservedRegex := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	filenameReservedWindowsNamesRegex := regexp.MustCompile(`(?i)^(con|prn|aux|nul|com[0-9]|lpt[0-9])$`)

	str = filenameReservedRegex.ReplaceAllString(str, "-")
	str = reControlCharsRegex.ReplaceAllString(str, "-")
	if filenameReservedWindowsNamesRegex.MatchString(str) {
		str = str + "!"
	}
	return str
}

func GetAltPath(path string) string {
	i := 1
	newPath := path
	for {
		_, e := os.Stat(newPath)
		if errors.Is(e, os.ErrNotExist) {
			return newPath
		}
		newPath = fmt.Sprintf("%s.%d", path, i)
		i += 1
	}
}
func GetFile(file string) string {
	b, e := os.ReadFile(file)
	if e == nil {
		return strings.TrimSpace(string(b))
	}
	return ""
}
func SetFile(file, val string) {
	_ = os.WriteFile(file, []byte(val), 0400)
}
func RenFile(src, dst string) error {
	e := os.MkdirAll(filepath.Dir(dst), 0755)
	if e != nil {
		return e
	}
	return os.Rename(src, dst)
}
func MvFile(src, dst string) error {
	e := os.MkdirAll(filepath.Dir(dst), 0755)
	if e != nil {
		return e
	}
	e = os.Rename(src, dst)
	if e == nil {
		return nil
	}
	if !strings.Contains(e.Error(), "invalid cross-device link") {
		return e
	}
	st, _ := os.Stat(src)
	in, e := os.Open(src)
	if e != nil {
		return e
	}
	out, e := os.Create(dst)
	if e != nil {
		return e
	}
	defer out.Close()
	_, e = io.Copy(out, in)
	if e != nil {
		return e
	}
	e = in.Close()
	if e != nil {
		return e
	}
	//e = out.Sync()
	//if e != nil {
	//	return e
	//}
	e = os.Chmod(dst, st.Mode())
	if e != nil {
		return e
	}
	e = os.Remove(src)
	if e != nil {
		return e
	}
	return nil
}

func MvTree(src, dst string, removeEmpties bool) {
	P("moving tree %s To %s", src, dst)
	var files []string
	var folders []string
	walk := func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(p, "_UNPACK_") || strings.Contains(p, "_FAILED_") {
			return nil
		}
		if info.IsDir() {
			folders = append(folders, p)
		} else {
			files = append(files, p)
		}
		return nil
	}
	err := filepath.Walk(src, walk)
	ChkFatal(err)

	for _, f := range folders {
		newFolder := strings.Replace(f, src, dst, 1)
		err := os.MkdirAll(newFolder, 0777)
		ChkFatal(err)
	}
	for _, f := range files {
		dstFile := strings.Replace(f, src, dst, 1)
		dstFile = GetAltPath(dstFile)
		P("moving file To %s", dstFile)
		renErr := os.Rename(f, dstFile)
		ChkFatal(renErr)

	}
	if removeEmpties {
		RmEmptyFolders(src)
	}
}

func FileExists(f string) bool {
	_, e := os.Stat(f)
	return e == nil
}

func GetTimestamp() string {
	return time.Now().Format("20060102-150105")
}

func LogOutput(logfile string) func() {
	var mwNoFile io.Writer

	// open file read/write | create if not exist | clear file at open if exists
	f, _ := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)

	// save existing stdout | MultiWriter writes to saved stdout and file
	out := os.Stdout
	mwNoFile = io.MultiWriter(out)
	mwFile := io.MultiWriter(f)
	mw := io.MultiWriter(mwFile, mwNoFile)

	// get pipe reader and writer | writes to pipe writer come out pipe reader
	r, w, _ := os.Pipe()

	// replace stdout,stderr with pipe writer | all writes to stdout, stderr will go through pipe instead (fmt.print, log)
	os.Stdout = w
	os.Stderr = w

	// writes with log.Print should also write to mw
	log.SetOutput(mw)

	//create channel to control exit | will block until all copies are finished
	exit := make(chan bool)

	go func() {
		// copy all reads from pipe to multiwriter, which writes to stdout and file
		_, _ = io.Copy(mw, r)
		// when r or w is closed copy will finish and true will be sent to channel
		exit <- true
	}()

	// function to be deferred in main until program exits
	return func() {
		// close writer then block on exit channel | this will let mw finish writing before the program exits
		_ = w.Close()
		<-exit
		// close file after all writes have finished
		_ = f.Close()
	}

}

func IsIp(host string) bool {
	re := regexp.MustCompile(`(\d{1,3}\.){3}\d{1,3}`)
	return re.MatchString(host)
}

func GetLocalIp() string {
	conn, err := net.Dial("udp", "255.255.255.255:65535")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
func DnsQueryServer(host, dnsServer string) []string {
	var ips []string
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, "udp", dnsServer+":53")
		},
	}
	ips, e := r.LookupHost(context.Background(), host)
	if e != nil && !strings.Contains(e.Error(), "no such host") {
		ChkFatal(e)
	}
	return ips
}
func DnsQueryServerA(host, dnsServer string) []string {
	return DnsQueryServer(host, dnsServer)
}

func DnsQueryServerPtr(ip, dnsServer string) []string {
	var hosts []string
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, "udp", dnsServer+":53")
		},
	}
	hosts, e := r.LookupAddr(context.Background(), ip)
	if e != nil && !strings.Contains(e.Error(), "no such host") {
		ChkFatal(e)
	}
	return hosts
}

func DnsQuery(host string) []string {
	var ips []string
	result, e := net.LookupIP(host)
	ChkFatal(e)
	for _, ip := range result {
		ips = append(ips, ip.String())
	}
	return ips
}

//func PStr(s string) *string {
//	return &s
//}
//func PInt(n int) *int {
//	return &n
//}
//func PBool(b bool) *bool {
//	return &b
//}
func Ptr[T any](a T) *T {
	return &a
}
