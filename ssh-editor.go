package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	rootPath   string
	useSudo    bool
	password   string
}

type FileNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"isDir"`
	Children []*FileNode `json:"children,omitempty"`
}

type ConnectRequest struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	UseSudo  bool   `json:"useSudo"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

var server *Server

func main() {
	server = &Server{}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/connect", handleConnect)
	http.HandleFunc("/api/tree", handleTree)
	http.HandleFunc("/api/file", handleFile)
	http.HandleFunc("/api/save", handleSave)
	http.HandleFunc("/api/create", handleCreate)
	http.HandleFunc("/api/delete", handleDelete)

	fmt.Println("🚀 SSH Code Editor démarré sur http://localhost:8080")
	fmt.Println("📝 Ouvrez votre navigateur et accédez à cette adresse")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SSH Code Editor Pro</title>
    <style>
        :root {
            --bg-primary: #1e1e1e;
            --bg-secondary: #252526;
            --bg-tertiary: #2d2d30;
            --bg-elevated: #333333;
            --border-color: #3e3e42;
            --border-focus: #007acc;
            --text-primary: #cccccc;
            --text-secondary: #969696;
            --text-muted: #6a6a6a;
            --accent: #007acc;
            --accent-hover: #1c97ea;
            --accent-dark: #005a9e;
            --danger: #f48771;
            --danger-hover: #f14c4c;
            --success: #89d185;
            --shadow-sm: 0 1px 3px rgba(0,0,0,0.4);
            --shadow-md: 0 2px 8px rgba(0,0,0,0.5);
            --shadow-lg: 0 4px 16px rgba(0,0,0,0.6);
        }
        
        * { 
            margin: 0; 
            padding: 0; 
            box-sizing: border-box; 
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Inter', sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            height: 100vh;
            overflow: hidden;
            font-size: 14px;
        }
        
        #app { 
            display: flex; 
            flex-direction: column; 
            height: 100vh; 
        }
        
        /* HEADER */
        #header {
            background: var(--bg-tertiary);
            padding: 8px 12px;
            border-bottom: 1px solid var(--border-color);
            display: flex;
            gap: 8px;
            align-items: center;
            height: 35px;
        }
        
        #header .spacer {
            flex: 1;
        }
        
        /* CONTENT AREA */
        #content {
            display: flex;
            flex: 1;
            overflow: hidden;
        }
        
        /* SIDEBAR */
        #sidebar {
            width: 250px;
            background: var(--bg-secondary);
            border-right: 1px solid var(--border-color);
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        
        #sidebar-header {
            padding: 8px 12px;
            background: var(--bg-secondary);
            border-bottom: 1px solid var(--border-color);
            display: flex;
            justify-content: space-between;
            align-items: center;
            font-size: 11px;
            text-transform: uppercase;
            font-weight: 600;
            letter-spacing: 0.5px;
            color: var(--text-secondary);
            height: 35px;
        }
        
        #sidebar-actions {
            display: flex;
            gap: 6px;
        }
        
        .icon-btn {
            background: transparent;
            border: none;
            color: var(--text-secondary);
            cursor: pointer;
            padding: 4px 6px;
            border-radius: 3px;
            font-size: 14px;
            transition: all 0.15s ease;
        }
        
        .icon-btn:hover {
            background: var(--bg-elevated);
            color: var(--text-primary);
        }
        
        #tree-container {
            flex: 1;
            overflow-y: auto;
            overflow-x: hidden;
            padding: 8px 0;
        }
        
        #tree-container::-webkit-scrollbar {
            width: 8px;
        }
        
        #tree-container::-webkit-scrollbar-track {
            background: transparent;
        }
        
        #tree-container::-webkit-scrollbar-thumb {
            background: var(--bg-elevated);
            border-radius: 4px;
        }
        
        #tree-container::-webkit-scrollbar-thumb:hover {
            background: var(--border-color);
        }
        
        /* TREE ITEMS */
        .tree-item {
            padding: 6px 12px;
            cursor: pointer;
            user-select: none;
            transition: background 0.12s ease;
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 13px;
            position: relative;
        }
        
        .tree-item:hover { 
            background: var(--bg-elevated);
        }
        
        .tree-item.selected { 
            background: var(--accent-dark);
            color: var(--text-primary);
        }
        
        .tree-item .icon {
            width: 16px;
            text-align: center;
            flex-shrink: 0;
            font-size: 12px;
        }
        
        .tree-item .name {
            flex: 1;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        
        .tree-item.folder .icon::before {
            content: '▶';
            display: inline-block;
            transition: transform 0.15s ease;
        }
        
        .tree-item.folder.expanded .icon::before {
            transform: rotate(90deg);
        }
        
        .tree-children {
            margin-left: 16px;
            border-left: 1px solid var(--border-color);
            padding-left: 4px;
        }
        
        /* EDITOR CONTAINER */
        #editor-container {
            flex: 1;
            display: flex;
            flex-direction: column;
            background: var(--bg-primary);
            overflow: hidden;
        }
        
        #editor-header {
            background: var(--bg-secondary);
            padding: 10px 20px;
            border-bottom: 1px solid var(--border-color);
            display: flex;
            justify-content: space-between;
            align-items: center;
            gap: 16px;
        }
        
        #editor-tabs {
            display: flex;
            gap: 4px;
            flex: 1;
            overflow-x: auto;
        }
        
        .editor-tab {
            background: var(--bg-tertiary);
            padding: 6px 14px;
            border-radius: 4px 4px 0 0;
            font-size: 13px;
            color: var(--text-secondary);
            cursor: pointer;
            transition: all 0.15s ease;
            white-space: nowrap;
            border: 1px solid transparent;
            border-bottom: none;
        }
        
        .editor-tab.active {
            background: var(--bg-primary);
            color: var(--text-primary);
            border-color: var(--border-color);
        }
        
        #editor-wrapper {
            flex: 1;
            display: flex;
            overflow: hidden;
            position: relative;
            background: var(--bg-primary);
        }
        
        #editor {
            width: 100%;
            height: 100%;
            background: var(--bg-primary);
            color: var(--text-primary);
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            font-size: 14px;
            padding: 20px;
            border: none;
            outline: none;
            resize: none;
            line-height: 1.5;
            tab-size: 4;
            overflow: auto;
        }
        
        #editor::selection {
            background: rgba(100, 150, 255, 0.3);
        }
        
        #editor::-webkit-scrollbar {
            width: 10px;
            height: 10px;
        }
        
        #editor::-webkit-scrollbar-track {
            background: var(--bg-primary);
        }
        
        #editor::-webkit-scrollbar-thumb {
            background: var(--bg-elevated);
            border-radius: 5px;
        }
        
        #editor::-webkit-scrollbar-thumb:hover {
            background: var(--border-color);
        }
        
        /* SYNTAX HIGHLIGHTING PREVIEW */
        #preview {
            position: absolute;
            top: 20px;
            left: 20px;
            right: 20px;
            pointer-events: none;
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            font-size: 13px;
            line-height: 1.6;
            white-space: pre;
            overflow: hidden;
        }
        
        /* STATUS BAR */
        #status-bar {
            background: var(--bg-secondary);
            color: var(--text-secondary);
            padding: 6px 20px;
            font-size: 12px;
            border-top: 1px solid var(--border-color);
            display: flex;
            justify-content: space-between;
            align-items: center;
            gap: 20px;
        }
        
        .status-item {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        
        /* BUTTONS */
        button {
            background: var(--bg-elevated);
            color: var(--text-primary);
            border: 1px solid transparent;
            padding: 5px 12px;
            border-radius: 2px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 400;
            transition: all 0.1s ease;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            height: 28px;
        }
        
        button:hover { 
            background: var(--bg-tertiary);
        }
        
        button:active {
            background: var(--bg-secondary);
        }
        
        button:disabled {
            background: var(--bg-secondary);
            color: var(--text-muted);
            cursor: not-allowed;
            opacity: 0.6;
        }
        
        button.primary {
            background: var(--accent);
            color: white;
        }
        
        button.primary:hover {
            background: var(--accent-hover);
        }
        
        button.primary:disabled {
            background: var(--accent-dark);
            opacity: 0.4;
        }
        
        button.danger {
            background: transparent;
            border-color: var(--danger);
            color: var(--danger);
        }
        
        button.danger:hover {
            background: var(--danger);
            color: white;
        }
        
        /* INPUTS */
        input, select {
            background: var(--bg-tertiary);
            color: var(--text-primary);
            border: 1px solid var(--border-color);
            padding: 10px 12px;
            border-radius: 4px;
            font-size: 13px;
            transition: all 0.2s ease;
            font-family: inherit;
        }
        
        input:focus, select:focus {
            outline: none;
            border-color: var(--border-focus);
            background: var(--bg-elevated);
            box-shadow: 0 0 0 3px rgba(13, 115, 119, 0.15);
        }
        
        input::placeholder {
            color: var(--text-muted);
        }
        
        /* MODAL */
        .modal {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.85);
            backdrop-filter: blur(8px);
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 1000;
            animation: fadeIn 0.2s ease;
        }
        
        @keyframes fadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
        }
        
        .modal.hidden { 
            display: none; 
        }
        
        .modal-content {
            background: var(--bg-secondary);
            padding: 28px;
            border-radius: 8px;
            width: 480px;
            max-width: 90%;
            border: 1px solid var(--border-color);
            box-shadow: var(--shadow-lg);
            animation: slideUp 0.25s ease;
        }
        
        @keyframes slideUp {
            from { 
                transform: translateY(30px);
                opacity: 0;
            }
            to { 
                transform: translateY(0);
                opacity: 1;
            }
        }
        
        .modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 24px;
        }
        
        .modal-header h2 {
            color: var(--text-primary);
            font-size: 20px;
            font-weight: 600;
        }
        
        .modal-close {
            background: transparent;
            border: none;
            color: var(--text-secondary);
            cursor: pointer;
            padding: 4px;
            font-size: 20px;
            line-height: 1;
        }
        
        .form-group {
            margin-bottom: 16px;
        }
        
        .form-group label {
            display: block;
            margin-bottom: 6px;
            font-size: 12px;
            color: var(--text-secondary);
            font-weight: 500;
            text-transform: uppercase;
            letter-spacing: 0.3px;
        }
        
        .form-group input, 
        .form-group select {
            width: 100%;
        }
        
        .form-group.checkbox {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        
        .form-group.checkbox input {
            width: auto;
        }
        
        .form-group.checkbox label {
            margin: 0;
            text-transform: none;
            font-size: 13px;
        }
        
        .form-buttons {
            display: flex;
            gap: 10px;
            margin-top: 24px;
        }
        
        .form-buttons button {
            flex: 1;
        }
        
        /* CONTEXT MENU */
        .context-menu {
            position: fixed;
            background: var(--bg-elevated);
            border: 1px solid var(--border-color);
            border-radius: 6px;
            padding: 4px 0;
            box-shadow: var(--shadow-lg);
            z-index: 10000;
            min-width: 180px;
        }
        
        .context-menu-item {
            padding: 8px 16px;
            cursor: pointer;
            font-size: 13px;
            display: flex;
            align-items: center;
            gap: 10px;
            transition: background 0.1s ease;
        }
        
        .context-menu-item:hover {
            background: var(--bg-tertiary);
        }
        
        .context-menu-item.danger {
            color: var(--danger);
        }
        
        .context-menu-divider {
            height: 1px;
            background: var(--border-color);
            margin: 4px 0;
        }
        
        /* LOADING */
        .loading {
            display: inline-block;
            width: 12px;
            height: 12px;
            border: 2px solid var(--border-color);
            border-top-color: var(--accent);
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
        }
        
        @keyframes spin {
            to { transform: rotate(360deg); }
        }
        
        /* NOTIFICATION */
        .notification {
            position: fixed;
            top: 20px;
            right: 20px;
            background: var(--bg-elevated);
            border: 1px solid var(--border-color);
            padding: 14px 18px;
            border-radius: 6px;
            box-shadow: var(--shadow-lg);
            z-index: 10000;
            animation: slideInRight 0.3s ease;
            max-width: 350px;
        }
        
        @keyframes slideInRight {
            from {
                transform: translateX(400px);
                opacity: 0;
            }
            to {
                transform: translateX(0);
                opacity: 1;
            }
        }
        
        .notification.success {
            border-left: 3px solid var(--success);
        }
        
        .notification.error {
            border-left: 3px solid var(--danger);
        }
    </style>
</head>
<body>
    <div id="app">
        <div id="header">
            <button onclick="showConnectModal()">Nouveau projet SSH</button>
            <button id="saveBtn" onclick="saveFile()" disabled class="primary">Sauvegarder</button>
            <div class="spacer"></div>
            <button onclick="disconnect()">Déconnecter</button>
        </div>
        
        <div id="content">
            <div id="sidebar">
                <div id="sidebar-header">
                    <span>Explorateur</span>
                    <div id="sidebar-actions">
                        <button class="icon-btn" onclick="showCreateModal('file')" title="Nouveau fichier">+</button>
                        <button class="icon-btn" onclick="showCreateModal('folder')" title="Nouveau dossier">□</button>
                        <button class="icon-btn" onclick="loadTree()" title="Rafraîchir">↻</button>
                    </div>
                </div>
                <div id="tree-container">
                    <div id="tree"></div>
                </div>
            </div>
            
            <div id="editor-container">
                <div id="editor-header">
                    <div id="editor-tabs">
                        <div class="editor-tab active">
                            <span id="current-file">Aucun fichier ouvert</span>
                        </div>
                    </div>
                    <span id="file-size" style="color: var(--text-muted); font-size: 12px;"></span>
                </div>
                <div id="editor-wrapper">
                    <textarea id="editor" placeholder="Sélectionnez un fichier pour commencer..." spellcheck="false"></textarea>
                </div>
            </div>
        </div>
        
        <div id="status-bar">
            <div class="status-item">
                <span id="status-text">Prêt</span>
            </div>
            <div class="status-item">
                <span id="connection-info"></span>
                <span id="language-info"></span>
            </div>
        </div>
    </div>

    <!-- Modal Connexion -->
    <div id="connectModal" class="modal hidden">
        <div class="modal-content">
            <div class="modal-header">
                <h2>Connexion SSH</h2>
                <button class="modal-close" onclick="hideConnectModal()">×</button>
            </div>
            <div class="form-group">
                <label>Type</label>
                <select id="type">
                    <option value="true">Dossier</option>
                    <option value="false">Fichier</option>
                </select>
            </div>
            <div class="form-group checkbox">
                <input type="checkbox" id="useSudo">
                <label for="useSudo">Utiliser sudo (fichiers système)</label>
            </div>
            <div class="form-group">
                <label>Hôte</label>
                <input type="text" id="host" placeholder="192.168.1.100">
            </div>
            <div class="form-group">
                <label>Port</label>
                <input type="text" id="port" value="22">
            </div>
            <div class="form-group">
                <label>Utilisateur</label>
                <input type="text" id="username" placeholder="root">
            </div>
            <div class="form-group">
                <label>Mot de passe</label>
                <input type="password" id="password">
            </div>
            <div class="form-group">
                <label>Chemin</label>
                <input type="text" id="path" placeholder="/root/project">
            </div>
            <div class="form-buttons">
                <button onclick="hideConnectModal()">Annuler</button>
                <button onclick="connect()" class="primary">Connecter</button>
            </div>
        </div>
    </div>

    <!-- Modal Création -->
    <div id="createModal" class="modal hidden">
        <div class="modal-content">
            <div class="modal-header">
                <h2 id="createTitle">Nouveau fichier</h2>
                <button class="modal-close" onclick="hideCreateModal()">×</button>
            </div>
            <div class="form-group">
                <label>Nom</label>
                <input type="text" id="createName" placeholder="fichier.txt">
            </div>
            <div class="form-buttons">
                <button onclick="hideCreateModal()">Annuler</button>
                <button onclick="createItem()" class="primary">Créer</button>
            </div>
        </div>
    </div>

    <script>
        let currentFile = '';
        let expandedFolders = new Set();
        let createType = 'file';
        let contextMenuTarget = null;

        // CONNEXION
        function showConnectModal() {
            document.getElementById('connectModal').classList.remove('hidden');
        }

        function hideConnectModal() {
            document.getElementById('connectModal').classList.add('hidden');
        }

        async function connect() {
            const data = {
                host: document.getElementById('host').value,
                port: document.getElementById('port').value,
                username: document.getElementById('username').value,
                password: document.getElementById('password').value,
                path: document.getElementById('path').value,
                isDir: document.getElementById('type').value === 'true',
                useSudo: document.getElementById('useSudo').checked
            };

            if (!data.host || !data.username || !data.password || !data.path) {
                showNotification('Tous les champs sont requis', 'error');
                return;
            }

            updateStatus('Connexion...', true);
            
            try {
                const res = await fetch('/api/connect', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(data)
                });
                const result = await res.json();
                
                if (result.success) {
                    hideConnectModal();
                    document.getElementById('connection-info').textContent = data.username + '@' + data.host;
                    showNotification('Connecté avec succès', 'success');
                    updateStatus('Connecté');
                    loadTree();
                } else {
                    showNotification(result.message, 'error');
                    updateStatus('Échec');
                }
            } catch (e) {
                showNotification('Erreur: ' + e.message, 'error');
                updateStatus('Erreur');
            }
        }

        // ARBORESCENCE
        async function loadTree() {
            updateStatus('Chargement...', true);
            const res = await fetch('/api/tree');
            const result = await res.json();
            if (result.success) {
                renderTree(result.data, document.getElementById('tree'));
                updateStatus('Prêt');
            } else {
                updateStatus('Erreur');
            }
        }

        function renderTree(nodes, container) {
            container.innerHTML = '';
            nodes.forEach(node => {
                renderNode(node, container, 0);
            });
        }

        function renderNode(node, container, level) {
            const div = document.createElement('div');
            div.className = 'tree-item ' + (node.isDir ? 'folder' : 'file');
            div.style.paddingLeft = (level * 16 + 12) + 'px';
            div.dataset.path = node.path;
            div.dataset.isDir = node.isDir;
            
            if (node.isDir) {
                div.classList.add(expandedFolders.has(node.path) ? 'expanded' : 'collapsed');
            }
            
            const icon = document.createElement('span');
            icon.className = 'icon';
            if (!node.isDir) {
                icon.textContent = getFileIcon(node.name);
            }
            div.appendChild(icon);
            
            const name = document.createElement('span');
            name.className = 'name';
            name.textContent = node.name;
            div.appendChild(name);
            
            div.onclick = (e) => {
                e.stopPropagation();
                if (node.isDir) {
                    toggleFolder(node.path);
                } else {
                    loadFile(node.path);
                }
            };
            
            div.oncontextmenu = (e) => {
                e.preventDefault();
                showContextMenu(e, node.path, node.isDir);
            };
            
            container.appendChild(div);
            
            if (node.isDir && node.children && expandedFolders.has(node.path)) {
                const childContainer = document.createElement('div');
                childContainer.className = 'tree-children';
                node.children.forEach(child => {
                    renderNode(child, childContainer, level + 1);
                });
                container.appendChild(childContainer);
            }
        }

        function toggleFolder(path) {
            if (expandedFolders.has(path)) {
                expandedFolders.delete(path);
            } else {
                expandedFolders.add(path);
            }
            loadTree();
        }

        function getFileIcon(filename) {
            const ext = filename.split('.').pop().toLowerCase();
            const icons = {
                'js': '≡', 'json': '{ }', 'html': '< >', 'css': '#',
                'py': 'py', 'go': 'go', 'java': 'j', 'php': 'php',
                'md': 'md', 'txt': 'txt', 'pdf': 'pdf', 'zip': 'zip',
                'jpg': 'img', 'png': 'img', 'gif': 'img', 'svg': 'svg',
                'mp3': '♪', 'mp4': '▶', 'sh': 'sh', 'log': 'log'
            };
            return icons[ext] || '•';
        }

        // FICHIERS
        async function loadFile(path) {
            updateStatus('Chargement...', true);
            
            try {
                const res = await fetch('/api/file?path=' + encodeURIComponent(path));
                const result = await res.json();
                
                if (result.success) {
                    currentFile = path;
                    const editor = document.getElementById('editor');
                    editor.value = result.data.content;
                    
                    document.getElementById('current-file').textContent = path.split('/').pop();
                    document.getElementById('file-size').textContent = formatBytes(result.data.size);
                    document.getElementById('saveBtn').disabled = false;
                    
                    const lang = detectLanguage(path);
                    document.getElementById('language-info').textContent = lang.toUpperCase();
                    
                    updateStatus(path);
                    
                    document.querySelectorAll('.tree-item').forEach(el => {
                        el.classList.remove('selected');
                        if (el.dataset.path === path) {
                            el.classList.add('selected');
                        }
                    });
                } else {
                    showNotification(result.message, 'error');
                }
            } catch (e) {
                showNotification('Erreur: ' + e.message, 'error');
            }
        }

        async function saveFile() {
            if (!currentFile) return;
            
            updateStatus('Sauvegarde...', true);
            
            try {
                const res = await fetch('/api/save', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({
                        path: currentFile,
                        content: document.getElementById('editor').value
                    })
                });
                const result = await res.json();
                
                if (result.success) {
                    showNotification('Fichier sauvegardé', 'success');
                    updateStatus('Sauvegardé');
                    setTimeout(() => updateStatus(currentFile), 2000);
                } else {
                    showNotification(result.message, 'error');
                }
            } catch (e) {
                showNotification('Erreur: ' + e.message, 'error');
            }
        }

        // CRÉATION
        function showCreateModal(type) {
            createType = type;
            document.getElementById('createTitle').textContent = type === 'file' ? 'Nouveau fichier' : 'Nouveau dossier';
            document.getElementById('createName').value = '';
            document.getElementById('createModal').classList.remove('hidden');
        }

        function hideCreateModal() {
            document.getElementById('createModal').classList.add('hidden');
        }

        async function createItem() {
            const name = document.getElementById('createName').value.trim();
            if (!name) {
                showNotification('Nom requis', 'error');
                return;
            }

            try {
                const res = await fetch('/api/create', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({
                        type: createType,
                        name: name
                    })
                });
                const result = await res.json();
                
                if (result.success) {
                    hideCreateModal();
                    showNotification('Créé avec succès', 'success');
                    loadTree();
                } else {
                    showNotification(result.message, 'error');
                }
            } catch (e) {
                showNotification('Erreur: ' + e.message, 'error');
            }
        }

        // MENU CONTEXTUEL
        function showContextMenu(e, path, isDir) {
            const existing = document.querySelector('.context-menu');
            if (existing) existing.remove();
            
            const menu = document.createElement('div');
            menu.className = 'context-menu';
            menu.style.left = e.pageX + 'px';
            menu.style.top = e.pageY + 'px';
            
            const item = document.createElement('div');
            item.className = 'context-menu-item';
            item.innerHTML = '<span>×</span> Supprimer';
            item.onclick = function() { deleteItem(path); };
            menu.appendChild(item);
            
            document.body.appendChild(menu);
            contextMenuTarget = path;
            
            setTimeout(() => {
                document.addEventListener('click', closeContextMenu);
            }, 10);
        }

        function closeContextMenu() {
            const menu = document.querySelector('.context-menu');
            if (menu) menu.remove();
            document.removeEventListener('click', closeContextMenu);
        }

        async function deleteItem(path) {
            if (!confirm('Supprimer ' + path + ' ?')) return;
            
            try {
                const res = await fetch('/api/delete', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({ path: path })
                });
                const result = await res.json();
                
                if (result.success) {
                    showNotification('Supprimé', 'success');
                    if (currentFile === path) {
                        currentFile = '';
                        document.getElementById('editor').value = '';
                        document.getElementById('saveBtn').disabled = true;
                    }
                    loadTree();
                } else {
                    showNotification(result.message, 'error');
                }
            } catch (e) {
                showNotification('Erreur: ' + e.message, 'error');
            }
        }

        // UTILITAIRES
        function disconnect() {
            currentFile = '';
            expandedFolders.clear();
            document.getElementById('editor').value = '';
            document.getElementById('tree').innerHTML = '';
            document.getElementById('current-file').textContent = 'Aucun fichier ouvert';
            document.getElementById('file-size').textContent = '';
            document.getElementById('connection-info').textContent = '';
            document.getElementById('language-info').textContent = '';
            document.getElementById('saveBtn').disabled = true;
            updateStatus('Déconnecté');
        }

        function updateStatus(msg, loading = false) {
            const statusEl = document.getElementById('status-text');
            statusEl.innerHTML = loading ? '<span class="loading"></span> ' + msg : msg;
        }

        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
        }

        function detectLanguage(path) {
            const ext = path.split('.').pop().toLowerCase();
            const langs = {
                'js': 'javascript', 'jsx': 'javascript',
                'ts': 'typescript', 'tsx': 'typescript',
                'json': 'json', 
                'html': 'html', 'htm': 'html', 
                'css': 'css', 'scss': 'scss',
                'py': 'python',
                'go': 'go',
                'java': 'java',
                'php': 'php',
                'rb': 'ruby',
                'rs': 'rust',
                'c': 'c',
                'cpp': 'cpp', 'cc': 'cpp', 'cxx': 'cpp',
                'cs': 'csharp',
                'kt': 'kotlin',
                'swift': 'swift',
                'sql': 'sql',
                'sh': 'bash', 'bash': 'bash',
                'yaml': 'yaml', 'yml': 'yaml',
                'xml': 'xml',
                'md': 'markdown'
            };
            return langs[ext] || 'plaintext';
        }
        
        function applySyntaxHighlighting(code, language) {
            const overlay = document.getElementById('editor-overlay');
            const editor = document.getElementById('editor');
            
            try {
                let highlighted;
                if (language === 'plaintext') {
                    highlighted = hljs.highlightAuto(code).value;
                } else {
                    highlighted = hljs.highlight(code, { language: language }).value;
                }
                overlay.innerHTML = highlighted;
            } catch (e) {
                try {
                    overlay.innerHTML = hljs.highlightAuto(code).value;
                } catch (e2) {
                    overlay.textContent = code;
                }
            }
        }

        function updateLineNumbers() {
            // Removed
        }

        function showNotification(message, type = 'success') {
            const notif = document.createElement('div');
            notif.className = 'notification ' + type;
            notif.textContent = message;
            document.body.appendChild(notif);
            
            setTimeout(() => {
                notif.style.animation = 'fadeOut 0.3s ease';
                setTimeout(() => notif.remove(), 300);
            }, 3000);
        }
        
        function syncScroll() {
            const editor = document.getElementById('editor');
            const overlay = document.getElementById('editor-overlay');
            
            const translateX = -editor.scrollLeft;
            const translateY = -editor.scrollTop;
            overlay.style.transform = 'translate(' + translateX + 'px, ' + translateY + 'px)';
        }
        
        function updateHighlightingOnEdit() {
            clearTimeout(window.highlightTimeout);
            window.highlightTimeout = setTimeout(() => {
                const editor = document.getElementById('editor');
                const lang = detectLanguage(currentFile);
                applySyntaxHighlighting(editor.value, lang);
            }, 300);
        }

        // ÉVÉNEMENTS
        document.getElementById('editor').addEventListener('input', function() {
            // No highlighting
        });
        
        document.getElementById('editor').addEventListener('keydown', (e) => {
            if (e.ctrlKey && e.key === 's') {
                e.preventDefault();
                saveFile();
            }
            if (e.key === 'Tab') {
                e.preventDefault();
                const start = e.target.selectionStart;
                const end = e.target.selectionEnd;
                e.target.value = e.target.value.substring(0, start) + '    ' + e.target.value.substring(end);
                e.target.selectionStart = e.target.selectionEnd = start + 4;
            }
        });

        document.getElementById('editor').addEventListener('scroll', function() {
            // No sync needed
        });

        window.onload = () => showConnectModal();
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, tmpl)
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	var req ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Requête invalide")
		return
	}

	config := &ssh.ClientConfig{
		User: req.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(req.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%s", req.Host, req.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		sendError(w, fmt.Sprintf("Connexion SSH échouée: %v", err))
		return
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		sendError(w, fmt.Sprintf("Connexion SFTP échouée: %v", err))
		return
	}

	server.sshClient = client
	server.sftpClient = sftpClient
	server.rootPath = req.Path
	server.useSudo = req.UseSudo
	server.password = req.Password

	sendSuccess(w, "Connecté avec succès", nil)
}

func handleTree(w http.ResponseWriter, r *http.Request) {
	if server.sftpClient == nil {
		sendError(w, "Non connecté")
		return
	}

	tree, err := buildTree(server.rootPath)
	if err != nil {
		sendError(w, fmt.Sprintf("Erreur: %v", err))
		return
	}

	sendSuccess(w, "", tree)
}

func buildTree(path string) ([]*FileNode, error) {
	info, err := server.sftpClient.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("impossible d'accéder: %v", err)
	}

	if !info.IsDir() {
		return []*FileNode{{
			Name:  filepath.Base(path),
			Path:  path,
			IsDir: false,
		}}, nil
	}

	entries, err := server.sftpClient.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("impossible de lire: %v", err)
	}

	var nodes []*FileNode
	var dirs, files []*FileNode

	for _, entry := range entries {
		fullPath := filepath.ToSlash(filepath.Join(path, entry.Name()))
		node := &FileNode{
			Name:  entry.Name(),
			Path:  fullPath,
			IsDir: entry.IsDir(),
		}

		if entry.IsDir() {
			children, err := buildTree(fullPath)
			if err == nil {
				node.Children = children
			}
			dirs = append(dirs, node)
		} else {
			files = append(files, node)
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	nodes = append(nodes, dirs...)
	nodes = append(nodes, files...)

	return nodes, nil
}

func handleFile(w http.ResponseWriter, r *http.Request) {
	if server.sftpClient == nil {
		sendError(w, "Non connecté")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		sendError(w, "Chemin requis")
		return
	}

	path = filepath.ToSlash(path)

	var content []byte
	var err error

	if server.useSudo {
		content, err = readFileWithSudo(path)
	} else {
		file, err2 := server.sftpClient.Open(path)
		if err2 != nil {
			sendError(w, fmt.Sprintf("Impossible d'ouvrir: %v", err2))
			return
		}
		defer file.Close()
		content, err = io.ReadAll(file)
	}

	if err != nil {
		sendError(w, fmt.Sprintf("Erreur de lecture: %v", err))
		return
	}

	data := map[string]interface{}{
		"content": string(content),
		"size":    len(content),
	}

	sendSuccess(w, "", data)
}

func handleSave(w http.ResponseWriter, r *http.Request) {
	if server.sftpClient == nil {
		sendError(w, "Non connecté")
		return
	}

	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Requête invalide")
		return
	}

	req.Path = filepath.ToSlash(req.Path)

	var err error
	if server.useSudo {
		err = writeFileWithSudo(req.Path, req.Content)
	} else {
		file, err2 := server.sftpClient.Create(req.Path)
		if err2 != nil {
			sendError(w, fmt.Sprintf("Impossible de créer: %v", err2))
			return
		}
		defer file.Close()
		_, err = io.WriteString(file, req.Content)
	}

	if err != nil {
		sendError(w, fmt.Sprintf("Erreur d'écriture: %v", err))
		return
	}

	sendSuccess(w, "Fichier sauvegardé", nil)
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	if server.sftpClient == nil {
		sendError(w, "Non connecté")
		return
	}

	var req struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Requête invalide")
		return
	}

	newPath := filepath.ToSlash(filepath.Join(server.rootPath, req.Name))

	if req.Type == "folder" {
		var err error
		if server.useSudo {
			err = createFolderWithSudo(newPath)
		} else {
			err = server.sftpClient.Mkdir(newPath)
		}
		if err != nil {
			sendError(w, fmt.Sprintf("Erreur: %v", err))
			return
		}
	} else {
		var err error
		if server.useSudo {
			err = writeFileWithSudo(newPath, "")
		} else {
			file, err2 := server.sftpClient.Create(newPath)
			if err2 != nil {
				sendError(w, fmt.Sprintf("Erreur: %v", err2))
				return
			}
			file.Close()
		}
		if err != nil {
			sendError(w, fmt.Sprintf("Erreur: %v", err))
			return
		}
	}

	sendSuccess(w, "Créé avec succès", nil)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if server.sftpClient == nil {
		sendError(w, "Non connecté")
		return
	}

	var req struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Requête invalide")
		return
	}

	req.Path = filepath.ToSlash(req.Path)

	var err error
	if server.useSudo {
		err = deleteWithSudo(req.Path)
	} else {
		info, err2 := server.sftpClient.Stat(req.Path)
		if err2 != nil {
			sendError(w, fmt.Sprintf("Erreur: %v", err2))
			return
		}
		if info.IsDir() {
			err = server.sftpClient.RemoveDirectory(req.Path)
		} else {
			err = server.sftpClient.Remove(req.Path)
		}
	}

	if err != nil {
		sendError(w, fmt.Sprintf("Erreur: %v", err))
		return
	}

	sendSuccess(w, "Supprimé", nil)
}

func sendSuccess(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func sendError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Message: message,
	})
}

func readFileWithSudo(path string) ([]byte, error) {
	session, err := server.sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	cmd := fmt.Sprintf("echo '%s' | sudo -S cat '%s'", server.password, path)
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 && strings.Contains(lines[0], "sudo") {
		return []byte(strings.Join(lines[1:], "\n")), nil
	}

	return output, nil
}

func writeFileWithSudo(path, content string) error {
	session, err := server.sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := fmt.Sprintf("echo '%s' | sudo -S tee '%s' > /dev/null", server.password, path)

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	if err := session.Start(cmd); err != nil {
		return err
	}

	io.WriteString(stdin, content)
	stdin.Close()

	return session.Wait()
}

func createFolderWithSudo(path string) error {
	session, err := server.sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := fmt.Sprintf("echo '%s' | sudo -S mkdir -p '%s'", server.password, path)
	return session.Run(cmd)
}

func deleteWithSudo(path string) error {
	session, err := server.sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := fmt.Sprintf("echo '%s' | sudo -S rm -rf '%s'", server.password, path)
	return session.Run(cmd)
}