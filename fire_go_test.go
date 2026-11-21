package main
import (
	"flag"
	"os"
	"strings"
	"testing"
)
func TestParseGitLogToTicker_NoOutput(t *testing.T) {
	if _, _, ok := parseGitLogToTicker(""); ok {
		t.Fatalf("ожидается ok=false для пустого журнала")
	}
}
func TestParseGitLogToTicker_SingleCommit(t *testing.T) {
	log := "abcd1234\tAlice\t3 дня назад\tПервоначальная фиксация"
	msg, meta, ok := parseGitLogToTicker(log)
	if !ok {
		t.Fatalf("ожидаемо ок=истина")
	}
	if msg == "" || meta == "" {
		t.Fatalf("ожидаемое непустое сообщение и мета")
	}
	if got, want := msg, "Первоначальная фиксация"; !contains(got, want) {
		t.Fatalf("сообщение %q не содержит %q", got, want)
	}
	if got, want := meta, "от Алиса 3 дня назад"; !contains(got, want) {
		t.Fatalf("мета %q не содержит %q", got, want)
	}
}
func TestParseGitLogToTicker_MultipleCommits(t *testing.T) {
	log := "" +
		"abcd1234\tAlice\t3 дня назад\tПервоначальная фиксация\n" +
		"efgh5678\tBob\t2 недели назад\tДобавить функцию X\n" +
		"ijkl9012\tCarol\t1 год назад\tРефакторинг модуля Y\n"
	msg, meta, ok := parseGitLogToTicker(log)
	if !ok {
		t.Fatalf("ожидаемо ок=истина")
	}
	for _, want := range []string{"Первоначальная фиксация", "Добавить функцию X", "Рефакторинг модуля Y"} {
		if !contains(msg, want) {
			t.Fatalf("Тикер сообщения %q не содержит %q", msg, want)
		}
	}
	for _, want := range []string{"от Алиса 3 дня назад", "от Боб 2 недели назад", "от Кэрол 1 год назад"} {
		if !contains(meta, want) {
			t.Fatalf("метатикер %q не содержит %q", meta, want)
		}
	}
}
func TestFlagParsing(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()
	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	contribs := flag.Bool("contribs", false, "Используйте визуализацию в виде диаграммы вклада GitHub.")
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		t.Fatalf("не удалось проанализировать флаги: %v", err)
	}
	if *contribs {
		t.Fatalf("ожидаемый вклад по умолчанию будет ложным")
	}
	os.Args = []string{"cmd", "--contribs"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	contribs = flag.Bool("contribs", false, "Используйте визуализацию в виде диаграммы вклада GitHub.")
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		t.Fatalf("не удалось проанализировать --contribs: %v", err)
	}
	if !*contribs {
		t.Fatalf("ожидаемый вклад будет истинным с флагом --contribs")
	}
	os.Args = []string{"cmd", "-contribs"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	contribs = flag.Bool("contribs", false, "Используйте визуализацию в виде диаграммы вклада GitHub.")
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		t.Fatalf("не удалось проанализировать -contribs: %v", err)
	}
	if !*contribs {
		t.Fatalf("ожидаемый вклад будет истинным с флагом -contribs")
	}
}
func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}