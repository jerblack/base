package progress

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func MvFile(src, dst string) error {
	//var mwNoFile io.Writer
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
	st, e := os.Stat(src)
	if e != nil {
		return e
	}
	//var bar *progressbar.ProgressBar
	//if st.Mode().IsRegular() {
	//	log.Println("is regular")
	//	bar = progressbar.NewOptions64(st.Size(),
	//		progressbar.OptionSetWriter(mwNoFile),
	//		progressbar.OptionSpinnerType(14),
	//		progressbar.OptionSetDescription(fmt.Sprintf("[bold][light_magenta] %s  [reset]", filepath.Base(dst))),
	//		progressbar.OptionShowBytes(true),
	//		progressbar.OptionSetPredictTime(false),
	//		progressbar.OptionShowCount(),
	//		progressbar.OptionClearOnFinish(),
	//		progressbar.OptionSetWidth(60),
	//		progressbar.OptionOnCompletion(func() {}),
	//		progressbar.OptionEnableColorCodes(true),
	//		progressbar.OptionThrottle(100*time.Millisecond),
	//		progressbar.OptionUseANSICodes(true),
	//		progressbar.OptionSetTheme(progressbar.Theme{
	//			Saucer:        "[magenta]█[reset]",
	//			SaucerHead:    "[light_magenta]█[reset]",
	//			SaucerPadding: "[_blue_] [reset]",
	//		}))
	//	bar.RenderBlank()
	//	log.Println(&bar)
	//} else {
	//	log.Println("is not regular")
	//	bar = nil
	//}

	in, e := os.Open(src)
	if e != nil {
		return e
	}
	out, e := os.Create(dst)
	if e != nil {
		return e
	}
	defer out.Close()
	//log.Println(bar)
	//log.Println(bar == nil)
	//log.Println(st.Mode())
	//log.Println(st.Mode().IsRegular())
	//log.Println(st.Mode().Type())

	//if bar != nil {
	//	log.Println(&bar)
	//	log.Println(bar.IsFinished())
	//	_, e = io.Copy(io.MultiWriter(out, bar), in)
	//} else {
	_, e = io.Copy(out, in)
	//}
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
	//if bar != nil {
	//	e = bar.Clear()
	//	if e != nil {
	//		return e
	//	}
	//}
	return nil
}
