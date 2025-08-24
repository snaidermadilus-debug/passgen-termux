package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"strings"
)

var (
	lower       = "abcdefghijklmnopqrstuvwxyz"
	upper       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits      = "0123456789"
	symbols     = "!@#$%^&*()-_=+[]{};:,.?/|~<>"
	ambiguous   = "O0oIl1|`'\";:,. "
)

type opts struct {
	length       int
	useLower     bool
	useUpper     bool
	useDigits    bool
	useSymbols   bool
	count        int
	exclude      string
	noAmbiguous  bool
	copyToClip   bool
}

func main() {
	o := parseFlags()

	charset := buildCharset(o)
	if len(charset) == 0 {
		fmt.Fprintln(os.Stderr, "Erro: nenhum conjunto de caracteres selecionado.")
		os.Exit(1)
	}

	for i := 0; i < o.count; i++ {
		pwd, err := generatePassword(o.length, charset, mustIncludes(o))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro ao gerar senha:", err)
			os.Exit(1)
		}
		fmt.Println(pwd)

		if o.copyToClip {
			if err := copyToClipboard(pwd); err != nil {
				fmt.Fprintln(os.Stderr, "Aviso: não consegui copiar para a área de transferência:", err)
			}
		}
	}
}

func parseFlags() opts {
	o := opts{}
	flag.IntVar(&o.length, "l", 16, "tamanho da senha")
	flag.BoolVar(&o.useLower, "lower", true, "incluir letras minúsculas")
	flag.BoolVar(&o.useUpper, "upper", true, "incluir letras maiúsculas")
	flag.BoolVar(&o.useDigits, "n", true, "incluir números")
	flag.BoolVar(&o.useSymbols, "s", false, "incluir símbolos")
	flag.IntVar(&o.count, "count", 1, "quantidade de senhas para gerar")
	flag.StringVar(&o.exclude, "exclude", "", "caracteres a excluir (ex: \"@0OIl| \")")
	flag.BoolVar(&o.noAmbiguous, "no-ambiguous", true, "remover caracteres ambíguos (ex: O,0,l,1,|)")
	flag.BoolVar(&o.copyToClip, "copy", false, "copiar a última senha para a área de transferência")
	flag.Parse()
	if o.length < 4 {
		o.length = 4
	}
	if o.count < 1 {
		o.count = 1
	}
	return o
}

func buildCharset(o opts) string {
	var sb strings.Builder
	if o.useLower {
		sb.WriteString(lower)
	}
	if o.useUpper {
		sb.WriteString(upper)
	}
	if o.useDigits {
		sb.WriteString(digits)
	}
	if o.useSymbols {
		sb.WriteString(symbols)
	}
	cs := uniqueRunes(sb.String())

	// Remover ambíguos e/ou excluídos
	if o.noAmbiguous {
		cs = removeRunes(cs, ambiguous)
	}
	if o.exclude != "" {
		cs = removeRunes(cs, o.exclude)
	}
	return cs
}

func mustIncludes(o opts) []string {
	sets := []string{}
	if o.useLower {
		sets = append(sets, lower)
	}
	if o.useUpper {
		sets = append(sets, upper)
	}
	if o.useDigits {
		sets = append(sets, digits)
	}
	if o.useSymbols {
		sets = append(sets, symbols)
	}
	// aplicar filtros de ambiguous/exclude nos conjuntos obrigatórios também
	if o.noAmbiguous {
		for i := range sets {
			sets[i] = removeRunes(sets[i], ambiguous)
		}
	}
	if o.exclude != "" {
		for i := range sets {
			sets[i] = removeRunes(sets[i], o.exclude)
		}
	}
	// descartar conjuntos vazios
	out := []string{}
	for _, s := range sets {
		if s != "" {
			out = append(out, uniqueRunes(s))
		}
	}
	return out
}

func generatePassword(length int, charset string, must []string) (string, error) {
	if len(charset) == 0 {
		return "", errors.New("charset vazio")
	}
	if length < len(must) {
		length = len(must)
	}

	// Garante pelo menos 1 de cada conjunto obrigatório
	pw := make([]rune, 0, length)
	for _, set := range must {
		r, err := pickRune(set)
		if err != nil {
			return "", err
		}
		pw = append(pw, r)
	}

	// Preenche o restante
	for len(pw) < length {
		r, err := pickRune(charset)
		if err != nil {
			return "", err
		}
		pw = append(pw, r)
	}

	// Embaralhar com Fisher–Yates usando crypto/rand
	if err := shuffleRunes(pw); err != nil {
		return "", err
	}

	return string(pw), nil
}

func pickRune(from string) (rune, error) {
	idx, err := cryptoRandInt(len(from))
	if err != nil {
		return 0, err
	}
	return []rune(from)[idx], nil
}

func shuffleRunes(runes []rune) error {
	for i := len(runes) - 1; i > 0; i-- {
		j, err := cryptoRandInt(i + 1)
		if err != nil {
			return err
		}
		runes[i], runes[j] = runes[j], runes[i]
	}
	return nil
}

func cryptoRandInt(n int) (int, error) {
	if n <= 0 {
		return 0, errors.New("n inválido")
	}
	max := big.NewInt(int64(n))
	v, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}

func uniqueRunes(s string) string {
	seen := map[rune]bool{}
	var out []rune
	for _, r := range s {
		if !seen[r] {
			seen[r] = true
			out = append(out, r)
		}
	}
	return string(out)
}

func removeRunes(s, remove string) string {
	rm := map[rune]bool{}
	for _, r := range remove {
		rm[r] = true
	}
	var out []rune
	for _, r := range s {
		if !rm[r] {
			out = append(out, r)
		}
	}
	return string(out)
}

func copyToClipboard(text string) error {
	// Termux
	if _, err := exec.LookPath("termux-clipboard-set"); err == nil {
		cmd := exec.Command("termux-clipboard-set")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}
	// X11/Linux comum
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command("xsel", "--clipboard", "--input")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}
	return errors.New("nenhuma ferramenta de clipboard encontrada (termux-clipboard-set/xclip/xsel)")
}
