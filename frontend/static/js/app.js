class NetworkScannerUI {
    constructor() {
        this.socket = null;
        this.pingInterval = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.baseReconnectDelay = 1000;
        this.currentTaskId = null;

        this.initElements();
        this.initEventListeners();
        this.initWebSocket();
    }

    initElements() {
        this.terminal = document.getElementById('terminal');
        this.scanButton = document.getElementById('scanButton');
        this.cancelButton = document.getElementById('cancelButton');
        this.ipInput = document.getElementById('ip');
        this.portsInput = document.getElementById('ports');
        this.progressContainer = document.getElementById('progressContainer');
        this.progressBar = document.getElementById('progressBar');
        this.progressText = document.getElementById('progressText');
        this.openPortsCount = document.getElementById('openPortsCount');
        this.openPortsList = document.getElementById('openPortsList');
        this.scanTarget = document.getElementById('scanTarget');
        this.connectionStatus = document.getElementById('connectionStatus');
    }

    initEventListeners() {
        this.scanButton.addEventListener('click', () => this.startScan());
        this.cancelButton.addEventListener('click', () => this.cancelScan());
    }

    initWebSocket() {
        if (this.socket) {
            this.socket.close();
        }

        const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
        const WS_URL = `ws://localhost:8080/ws`;

        this.addTerminalLine('Connecting to WebSocket server...', 'info-line');
        this.socket = new WebSocket(WS_URL);

        this.socket.onopen = (e) => {
            this.addTerminalLine('Successfully connected to server', 'success-line');
            this.reconnectAttempts = 0;
            this.updateConnectionStatus(true);

            this.pingInterval = setInterval(() => {
                if (this.socket.readyState === WebSocket.OPEN) {
                    this.socket.send(JSON.stringify({ action: "ping" }));
                }
            }, 25000);
        };

        this.socket.onclose = (e) => {
            this.addTerminalLine(`Connection closed: ${e.code} ${e.reason || ''}`, 'error-line');
            clearInterval(this.pingInterval);
            this.updateConnectionStatus(false);

            if (this.reconnectAttempts < this.maxReconnectAttempts) {
                const delay = this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts);
                this.reconnectAttempts++;
                this.addTerminalLine(`Reconnecting in ${delay/1000} seconds... (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`, 'info-line');
                setTimeout(() => this.initWebSocket(), delay);
            }
        };

        this.socket.onerror = (error) => {
            this.addTerminalLine('WebSocket error occurred', 'error-line');
            this.updateConnectionStatus(false);
        };

        this.socket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                
                if (message.type === "pong") return;

                switch (message.type) {
                    case 'welcome':
                        this.addTerminalLine(message.data?.message || 'Connected to scanner service', 'info-line');
                        break;
                    case 'scan_started':
                        this.handleScanStarted(message.data);
                        break;
                    case 'scan_queued':
                        this.handleScanQueued(message.data);
                        break;
                    case 'scan_result':
                        this.handleScanResult(message.data);
                        break;
                    case 'scan_progress':
                        const progress = Math.round(message.data.progress * 100);
                        this.progressBar.style.width = `${progress}%`;
                        this.progressText.textContent = `${progress}%`;
                        break;
                    case 'error':
                        this.addTerminalLine(`Error: ${message.message}`, 'error-line');
                        break;
                    default:
                        this.addTerminalLine(`Unknown message type: ${message.type}`, 'error-line');
                }
            } catch (e) {
                this.addTerminalLine('Error parsing message: ' + e.message, 'error-line');
            }
        };
    }

    addTerminalLine(text, className = '') {
        const line = document.createElement('div');
        line.className = `terminal-line ${className}`;
        line.textContent = `[${new Date().toLocaleTimeString()}] ${text}`;
        this.terminal.appendChild(line);
        this.terminal.scrollTop = this.terminal.scrollHeight;
    }

    updateConnectionStatus(connected) {
        this.connectionStatus.textContent = connected ? 'Connected' : 'Disconnected';
        this.connectionStatus.style.backgroundColor = connected ? '#4CAF50' : '#f44336';
    }

    startScan() {
        const ip = this.ipInput.value.trim();
        const ports = this.portsInput.value.trim();

        if (!ip) {
            alert('Please enter IP address');
            return;
        }

        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify({
                action: 'scan',
                data: { ip, ports }
            }));
        } else {
            this.addTerminalLine('Not connected to server', 'error-line');
        }
    }

    cancelScan() {
        this.resetScanUI();
        this.addTerminalLine('Scan canceled by user', 'info-line');
    }

    resetScanUI() {
        this.progressContainer.style.display = 'none';
        this.scanButton.style.display = 'inline-block';
        this.cancelButton.style.display = 'none';
        this.currentTaskId = null;
    }

    handleScanStarted(data) {
        this.scanTarget.textContent = `${data.ip || 'N/A'} (${data.ports || 'N/A'})`;
        this.progressContainer.style.display = 'block';
        this.scanButton.style.display = 'none';
        this.cancelButton.style.display = 'inline-block';
        this.progressBar.style.width = '0%';
        this.progressText.textContent = '0%';
        this.openPortsCount.textContent = '0';
        this.openPortsList.innerHTML = '';

        this.addTerminalLine(`Scan started for ${data.ip || 'N/A'} (ports: ${data.ports || 'N/A'})`, 'info-line');
    }

    handleScanQueued(data) {
        this.currentTaskId = data.task_id;
        this.addTerminalLine(`Scan queued with task ID: ${data.task_id}`, 'info-line');
    }

    handleScanResult(data) {
        const totalPorts = data.ports.includes('-') ?
            parseInt(data.ports.split('-')[1]) - parseInt(data.ports.split('-')[0]) + 1 :
            data.ports.split(',').length;

        const scanInfo = `
=== Scan Results ===
Target: ${data.ip}
Port range: ${data.ports}
Scan started: ${new Date(data.timestamp).toLocaleString()}
Duration: ${data.duration.toFixed(2)} seconds
Total ports scanned: ${totalPorts}
Open ports found: ${data.count}
Open ports list: ${data.open_ports.join(', ')}
===================
`;

        this.addTerminalLine(scanInfo, 'success-line');

        this.progressBar.style.width = '100%';
        this.progressText.textContent = '100%';
        this.openPortsCount.textContent = data.count;

        this.openPortsList.innerHTML = '';
        data.open_ports.forEach(port => {
            const portElement = document.createElement('span');
            portElement.className = 'port';
            portElement.textContent = port;
            this.openPortsList.appendChild(portElement);
        });

        this.addTerminalLine(`Scan completed successfully!`, 'success-line');
        this.resetScanUI();
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new NetworkScannerUI();
});