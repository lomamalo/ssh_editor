# SSH Code Editor

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

Éditeur de code professionnel accessible via navigateur pour éditer des fichiers distants via SSH/SFTP. Interface moderne inspirée de VSCode.

## ✨ Fonctionnalités

- **Interface web** moderne et intuitive
- **Connexion SSH/SFTP** sécurisée
- **Explorateur de fichiers** avec arborescence complète
- **Éditeur de code** avec support multi-langages
- **Sauvegarde** en temps réel (Ctrl+S)
- **Support sudo** pour éditer les fichiers système
- **Création/suppression** de fichiers et dossiers
- **Thème sombre** professionnel
- **Léger et rapide** - un seul binaire, aucune dépendance


## Installation

### Prérequis

- Go 1.21 ou supérieur (pour compiler depuis les sources)
- Accès SSH à un serveur distant

### Compilation depuis les sources
```bash
# Cloner le dépôt
git clone https://github.com/votre-username/ssh-code-editor.git
cd ssh-code-editor

# Installer les dépendances
go mod tidy

# Compiler
go build -o ssh-editor
```

### Compilation multi-plateforme

#### Windows (PowerShell)
```powershell
# Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o ssh-editor-linux

# macOS Intel
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o ssh-editor-macos

# macOS ARM
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o ssh-editor-macos-arm
```

#### Linux/macOS (Bash)
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o ssh-editor-linux

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o ssh-editor-macos

# macOS ARM
GOOS=darwin GOARCH=arm64 go build -o ssh-editor-macos-arm
```

## 📖 Utilisation

### Démarrage

#### Windows
```cmd
ssh-editor.exe
```

#### Linux/macOS
```bash
# Rendre exécutable (première fois seulement)
chmod +x ssh-editor-linux

# Lancer
./ssh-editor-linux
```

### Accès

1. Ouvrez votre navigateur sur **http://localhost:8080**
2. Une fenêtre de connexion SSH s'affiche automatiquement

### Connexion SSH

Remplissez les informations :

- **Type** : Dossier ou Fichier unique
- **Hôte** : IP du serveur (ex: `192.168.1.100`)
- **Port** : Port SSH (défaut: `22`)
- **Utilisateur** : Nom d'utilisateur SSH
- **Mot de passe** : Mot de passe SSH
- **Chemin** : Chemin du dossier/fichier (ex: `/home/user/project`)
- **Utiliser sudo** : Cocher si vous éditez des fichiers système

### Raccourcis clavier

- **Ctrl + S** : Sauvegarder le fichier
- **Tab** : Indentation (4 espaces)
- **Clic droit** : Menu contextuel (supprimer)

### Fonctionnalités avancées

#### Créer un fichier/dossier
Cliquez sur les boutons `+` ou `□` dans l'en-tête de l'explorateur.

#### Supprimer un fichier/dossier
Clic droit sur l'élément → Supprimer

#### Édition avec sudo
Cochez "Utiliser sudo" lors de la connexion pour éditer les fichiers système nécessitant des privilèges root.

## Dépendances Go
```go
github.com/pkg/sftp
golang.org/x/crypto/ssh
```

## Sécurité

**Important** :

- Les mots de passe SSH sont stockés temporairement en mémoire
- Utilisez HTTPS en production (reverse proxy recommandé)
- Limitez l'accès au port 8080 via firewall
- N'exposez pas directement sur Internet sans authentification supplémentaire