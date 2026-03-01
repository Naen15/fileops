package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fileops/internal/cfg"
	"fileops/internal/infra"
	"fileops/internal/ops"
	"fileops/internal/proc"
	"fileops/internal/secure"
	"fileops/internal/wiki"
)

func main() {
	flag.String("config", "config.json", "chemin du config.json")
	flag.Parse()

	conf, err := cfg.Load()
	if err != nil {
		log.Fatalf("Config: %v\n", err)
	}

	currentFile := conf.DefaultFile
	in := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf(`
============= MENU ===============
[f] Changer de fichier courant (actuel: %s)
[a] Analyse fichier courant
[b] Analyse répertoire
[c] Analyser une page Wikipédia
[d] ProcessOps (lister, filtrer, kill)
[e] SecureOps (verrou, read-only, audit)
[g] ContainerOps  (Docker ps, stats)
[q] Quitter
> `, currentFile)

		if !in.Scan() {
			fmt.Println("\nBye.")
			return
		}
		choice := in.Text()

		switch choice {
		case "f":
			fmt.Print("Chemin du fichier : ")
			if !in.Scan() {
				continue
			}
			path := in.Text()
			if path == "" {
				path = currentFile
			}
			if info, err := os.Stat(path); err != nil || info.IsDir() {
				fmt.Println("Fichier invalide")
			} else {
				currentFile = path
			}

		case "a":
			if err := runSingleFile(conf, currentFile); err != nil {
				fmt.Printf("Erreur: %v\n", err)
			}

		case "b":
			fmt.Print("Répertoire : ")
			if !in.Scan() {
				continue
			}
			dir := in.Text()
			if dir == "" {
				dir = conf.BaseDir
			}
			if err := runBatch(conf, dir); err != nil {
				fmt.Printf("Erreur: %v\n", err)
			}

		case "c":
			fmt.Print("Article(s) Wikipédia (séparés par ,) : ")
			if !in.Scan() {
				break
			}
			raw := in.Text()
			titles := strings.Split(raw, ",")

			arts, err := wiki.FetchMany(titles)
			if err != nil {
				fmt.Println("Erreur :", err)
				break
			}
			for _, a := range arts {
				if path, err := wiki.Save(a, conf.OutDir); err == nil {
					fmt.Printf("OK  %s → %s  (%d mots)\n", a.Title, path, a.Words)
				}
			}

		case "d":
			processOps()

		case "e":
			secureMenu(conf)

		case "g":
			containerMenu()
		case "q":
			fmt.Println("À la prochaine")
			return
		default:
			fmt.Println("Choix inconnu.")
		}
	}
}

func runSingleFile(conf cfg.Config, path string) error {
	lines, err := ops.ReadLines(path)
	if err != nil {
		return err
	}

	if err := ops.PrintFileInfo(path, lines); err != nil {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Mot-clé pour filtrer : ")
	if !scanner.Scan() {
		return nil
	}
	keyword := scanner.Text()

	if err := ops.FilterLines(lines, keyword, filepath.Join(conf.OutDir, "filtered.txt"), true); err != nil {
		return err
	}
	if err := ops.FilterLines(lines, keyword, filepath.Join(conf.OutDir, "filtered_not.txt"), false); err != nil {
		return err
	}

	fmt.Print("Combien de lignes pour head/tail ? ")
	if !scanner.Scan() {
		return nil
	}
	var n int
	fmt.Sscan(scanner.Text(), &n)

	_ = ops.WriteLines(lines[:min(n, len(lines))], filepath.Join(conf.OutDir, "head.txt"))
	_ = ops.WriteLines(lines[max(0, len(lines)-n):], filepath.Join(conf.OutDir, "tail.txt"))

	return nil
}

func runBatch(conf cfg.Config, dir string) error {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = conf.BaseDir
	}

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return fmt.Errorf("répertoire invalide : %v", err)
	}

	files, err := ops.ListTxt(dir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Printf("Aucun fichier .txt trouvé dans %s\n", dir)
		return nil
	}

	report := filepath.Join(conf.OutDir, "report.txt")
	index := filepath.Join(conf.OutDir, "index.txt")
	merged := filepath.Join(conf.OutDir, "merged.txt")

	if err := ops.ProcessBatch(files, report, index, merged); err != nil {
		return err
	}
	fmt.Printf("Analyse terminée : %d fichiers .txt → résultats dans %s\n",
		len(files), conf.OutDir)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func processOps() {
	in := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(`
----- ProcessOps -----
[1] Lister les 10 premiers processus
[2] Rechercher un mot-clé
[3] Kill (avec confirmation)
[z] Retour
> `)
		if !in.Scan() {
			return
		}
		switch strings.TrimSpace(in.Text()) {
		case "1":
			if list, err := proc.List(); err != nil {
				fmt.Println("Erreur :", err)
			} else {
				// en-tête pour faire la part des données des processus
				fmt.Printf("\n%-6s  %s\n", "PID", "Processus")
				fmt.Println(strings.Repeat("-", 25))

				for i, p := range list {
					if i >= 10 {
						break
					}
					fmt.Printf("%-6d  %s\n", p.PID, p.Name)
				}
			}

		case "2":
			fmt.Print("Mot-clé : ")
			if !in.Scan() {
				continue
			}
			kw := strings.TrimSpace(in.Text())
			all, err := proc.List()
			if err != nil {
				fmt.Println("Erreur :", err)
				continue
			}
			for _, p := range proc.Filter(all, kw) {
				fmt.Printf("%5d  %s\n", p.PID, p.Name)
			}

		case "3":
			fmt.Print("PID à tuer : ")
			if !in.Scan() {
				continue
			}
			pidStr := strings.TrimSpace(in.Text())
			pid, err := strconv.Atoi(pidStr)
			if err != nil || pid <= 0 {
				fmt.Println("PID invalide")
				continue
			}

			name := "(inconnu)"
			if list, _ := proc.List(); list != nil {
				for _, p := range list {
					if p.PID == pid {
						name = p.Name
						break
					}
				}
			}
			fmt.Printf("Confirmer kill de %s (PID %d) ? yes/no : ", name, pid)
			if !in.Scan() {
				continue
			}
			if strings.ToLower(strings.TrimSpace(in.Text())) != "yes" {
				fmt.Println("Annulé.")
				continue
			}
			if err := proc.Kill(pid, false); err != nil {
				fmt.Println("Erreur :", err)
			} else {
				fmt.Println("Processus terminé (ou déjà mort).")
			}

		case "z":
			return
		default:
			fmt.Println("Choix inconnu.")
		}
	}
}

func secureMenu(conf cfg.Config) {
	in := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(`
----- SecureOps -----
[1] Verrouiller un fichier
[2] Déverrouiller un fichier
[3] Rendre read-only
[z] Retour
> `)
		if !in.Scan() {
			return
		}
		switch strings.TrimSpace(in.Text()) {
		case "1":
			fmt.Print("Fichier à verrouiller : ")
			if !in.Scan() {
				continue
			}
			file := strings.TrimSpace(in.Text())
			if file == "" {
				fmt.Println("Chemin vide.")
				continue
			}
			if lock, err := secure.Lock(file, conf.OutDir); err != nil {
				fmt.Println("Erreur :", err)
			} else {
				fmt.Println("Lock créé :", lock)
				secure.Log(conf.OutDir, "LOCK", lock)
			}
		case "2":
			fmt.Print("Fichier à déverrouiller : ")
			if !in.Scan() {
				continue
			}
			file := strings.TrimSpace(in.Text())
			lock, err := secure.Unlock(file, conf.OutDir)
			if err != nil {
				fmt.Println("Erreur :", err)
			} else {
				fmt.Println("Lock supprimé :", lock)
				secure.Log(conf.OutDir, "UNLOCK", lock)
			}
		case "3":
			fmt.Print("Fichier à passer en read-only : ")
			if !in.Scan() {
				continue
			}
			file := strings.TrimSpace(in.Text())
			if err := secure.MakeReadOnly(file); err != nil {
				fmt.Println("Erreur :", err)
			} else {
				fmt.Println("Mode lecture-seule appliqué.")
				secure.Log(conf.OutDir, "CHMOD RO", file)
			}
		case "z":
			return
		default:
			fmt.Println("Choix inconnu.")
		}
	}
}
func containerMenu() {
	in := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(`
----- ContainerOps -----
[1] Lister les conteneurs actifs
[2] Stats CPU/Mem d’un conteneur
[z] Retour
> `)
		if !in.Scan() {
			return
		}
		switch strings.TrimSpace(in.Text()) {
		case "1":
			cs, err := infra.List()
			if err != nil {
				fmt.Println("Erreur :", err)
				return
			}
			if len(cs) == 0 {
				fmt.Println("Aucun conteneur en cours.")
				continue
			}
			fmt.Printf("\n%-12s %-20s %-20s %s\n", "ID", "NOM", "IMAGE", "STATUT")
			for _, c := range cs {
				fmt.Printf("%-12s %-20s %-20s %s\n",
					c.ID[:12], c.Name, c.Image, c.Status)
			}
		case "2":
			fmt.Print("ID ou nom du conteneur : ")
			if !in.Scan() {
				continue
			}
			id := strings.TrimSpace(in.Text())
			if id == "" {
				continue
			}
			if stat, err := infra.Stats(id); err != nil {
				fmt.Println("Erreur :", err)
			} else {
				fmt.Println("CPU%  MEM% :", stat)
			}
		case "z":
			return
		default:
			fmt.Println("Choix inconnu.")
		}
	}
}
