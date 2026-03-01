## fileops

Application CLI en Go qui regroupe :

• analyse de fichiers texte

• téléchargement d’articles Wikipédia

• gestion des processus (macOS et Windows)

• verrouillage, lecture-seule de fichiers et journalisation

• infos Docker (ps et stats)

• alerte espace disque (< 10 %)

## Installation rapide

git clone https://github.com/Naen15/fileops.git
cd fileops
go mod tidy # récupère les dépendances
go run ./cmd/fileops # lance le menu

ou

go build -o fileops ./cmd/fileops
./fileops --config mon_config.json # JSON facultatif

## Menu principal

[a] Analyse fichier

[b] Batch répertoire (.txt) (parallèle)

[c] WikiOps (1 ou n articles) (parallèle)

[d] ProcessOps (list, kill sécurisés)

[e] SecureOps (lock, read-only, audit)

[g] ContainerOps (docker ps et stats)

[q] Quitter
