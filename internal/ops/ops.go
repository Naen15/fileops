package ops

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func PrintFileInfo(path string, lines []string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	words, avgLen := statsWords(lines)

	fmt.Printf("\n— Infos sur %s —\n", path)
	fmt.Printf("Taille : %d o\n", info.Size())
	fmt.Printf("Créé : %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
	fmt.Printf("Nb lignes : %d\n", len(lines))
	fmt.Printf("Nb mots : %d (longueur moyenne %.1f)\n\n", words, avgLen)
	return nil
}

func FilterLines(lines []string, kw, out string, keep bool) error {
	var filtered []string
	for _, l := range lines {
		has := strings.Contains(strings.ToLower(l), strings.ToLower(kw))
		if (keep && has) || (!keep && !has) {
			filtered = append(filtered, l)
		}
	}
	return WriteLines(filtered, out)
}

func WriteLines(lines []string, out string) error {
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, l := range lines {
		if _, err := w.WriteString(l + "\n"); err != nil {
			return err
		}
	}
	return w.Flush()
}

func ListTxt(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".txt") {
			files = append(files, p)
		}
		return nil
	})
	return files, err
}

func ProcessBatch(files []string, report, index, merged string) error {
	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		indexLines  []string
		reportLines []string
		mergedLines []string
	)

	for _, f := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			info, err := os.Stat(file)
			if err != nil {
				return
			}

			lines, err := ReadLines(file)
			if err != nil {
				return
			}
			words, avg := statsWords(lines)

			mu.Lock()
			indexLines = append(indexLines,
				fmt.Sprintf("%s | %d o | %s",
					file, info.Size(),
					info.ModTime().Format("2006-01-02 15:04:05")))
			reportLines = append(reportLines,
				fmt.Sprintf("%s → %d lignes, %d mots (moy. %.1f)",
					filepath.Base(file), len(lines), words, avg))
			mergedLines = append(mergedLines, lines...)
			mu.Unlock()
		}(f)
	}

	wg.Wait()

	if err := WriteLines(indexLines, index); err != nil {
		return err
	}
	if err := WriteLines(reportLines, report); err != nil {
		return err
	}
	return WriteLines(mergedLines, merged)
}

func statsWords(lines []string) (int, float64) {
	var count, sum int
	for _, l := range lines {
		for _, tok := range strings.Fields(l) {
			if unicode.IsDigit(rune(tok[0])) {
				continue
			}
			count++
			sum += len(tok)
		}
	}
	if count == 0 {
		return 0, 0
	}
	return count, float64(sum) / float64(count)
}
