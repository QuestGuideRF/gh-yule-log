package main
import (
	"flag"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"github.com/gdamore/tcell/v2"
)
func parseGitLogToTicker(logOutput string) (string, string, bool) {
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	var msgSegs, metaSegs []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) != 4 {
			continue
		}
		_, author, relTime, subject := parts[0], parts[1], parts[2], parts[3]
		message := subject
		meta := "by " + author + " " + relTime
		msgRunes := []rune(message)
		metaRunes := []rune(meta)
		segmentWidth := len(msgRunes)
		if len(metaRunes) > segmentWidth {
			segmentWidth = len(metaRunes)
		}
		segmentWidth += 4
		msgSegs = append(msgSegs, padRight(message, segmentWidth))
		metaSegs = append(metaSegs, padRight(meta, segmentWidth))
	}
	if len(msgSegs) == 0 {
		return "", "", false
	}
	return strings.Join(msgSegs, ""), strings.Join(metaSegs, ""), true
}
func padRight(s string, n int) string {
	rs := []rune(s)
	if len(rs) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(rs))
}
func buildGitTickerText(maxCommits int) (string, string, bool) {
	args := []string{
		"log",
		"-n", strconv.Itoa(maxCommits),
		"--pretty=формат:%h%x09%an%x09%ar%x09%s",
	}
	cmd := exec.Command("git", args...)
	if dir := os.Getenv("YULE_LOG_GIT_DIR"); dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", "", false
	}
	return parseGitLogToTicker(string(out))
}
func main() {
	contribs := flag.Bool("contribs", false, "Используйте визуализацию в виде диаграммы вклада GitHub.")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("экран создания: %v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("экран инициализации: %v", err)
	}
	defer s.Fini()
	s.Clear()
	s.HideCursor()
	width, height := s.Size()
	if width <= 0 || height <= 0 {
		return
	}
	size := width * height
	buffer := make([]int, size+width+1)
	var chars []rune
	var styles []tcell.Style
	if *contribs {
		chars = []rune{' ', '⬝', '⬝', '⯀', '⯀', '◼', '◼', '■', '■', '■'}
		styles = []tcell.Style{
			tcell.StyleDefault.Foreground(tcell.ColorBlack),
			tcell.StyleDefault.Foreground(tcell.NewRGBColor(155, 233, 168)),
			tcell.StyleDefault.Foreground(tcell.NewRGBColor(64, 196, 99)),
			tcell.StyleDefault.Foreground(tcell.NewRGBColor(48, 161, 78)),
			tcell.StyleDefault.Foreground(tcell.NewRGBColor(33, 110, 57)),
		}
	} else {
		chars = []rune{' ', '.', ':', '^', '*', 'x', 's', 'S', '#', '$'}
		styles = []tcell.Style{
			tcell.StyleDefault.Foreground(tcell.ColorBlack),
			tcell.StyleDefault.Foreground(tcell.ColorMaroon),
			tcell.StyleDefault.Foreground(tcell.ColorRed),
			tcell.StyleDefault.Foreground(tcell.ColorDarkOrange),
			tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
		}
	}
	msgText, metaText, haveTicker := buildGitTickerText(20)
	msgRow := height - 2
	metaRow := height - 1
	tickerOffset := 0
	frame := 0
	events := make(chan tcell.Event, 10)
	go func() {
		for {
			ev := s.PollEvent()
			if ev == nil {
				return
			}
			events <- ev
		}
	}()
	frameDelay := 30 * time.Millisecond
loop:
	for {
		select {
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				_ = ev
				break loop
			case *tcell.EventResize:
				width, height = s.Size()
				if width <= 0 || height <= 0 {
					break loop
				}
				size = width * height
				buffer = make([]int, size+width+1)
				msgRow = height - 2
				metaRow = height - 1
			}
		default:
		}
		for i := 0; i < width/9; i++ {
			idx := rand.Intn(width) + width*(height-1)
			if idx >= 0 && idx < len(buffer) {
				buffer[idx] = 65
			}
		}
		for i := 0; i < size; i++ {
			b0 := buffer[i]
			b1 := buffer[i+1]
			b2 := buffer[i+width]
			b3 := buffer[i+width+1]
			v := (b0 + b1 + b2 + b3) / 4
			buffer[i] = v
			row := i / width
			col := i % width
			if row >= height || col >= width {
				continue
			}
			if haveTicker && row >= height-2 {
				continue
			}
			var style tcell.Style
			switch {
			case v > 15:
				style = styles[4]
			case v > 9:
				style = styles[3]
			case v > 4:
				style = styles[2]
			default:
				style = styles[1]
			}
			chIdx := v
			if chIdx > 9 {
				chIdx = 9
			}
			if chIdx < 0 {
				chIdx = 0
			}
			s.SetContent(col, row, chars[chIdx], nil, style)
		}
		if haveTicker && height >= 2 && len(msgText) > 0 {
			msgRunes := []rune(msgText)
			metaRunes := []rune(metaText)
			msgLen := len(msgRunes)
			metaLen := len(metaRunes)
			if msgLen > 0 && metaLen > 0 {
				for x := 0; x < width; x++ {
					mi := (tickerOffset + x) % msgLen
					mj := (tickerOffset + x) % metaLen
					mr := msgRunes[mi]
					me := metaRunes[mj]
					s.SetContent(x, msgRow, mr, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
					s.SetContent(x, metaRow, me, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
				}
				if frame%4 == 0 {
					tickerOffset = (tickerOffset + 1) % msgLen
				}
			}
		}
		s.Show()
		time.Sleep(frameDelay)
		frame++
	}
}