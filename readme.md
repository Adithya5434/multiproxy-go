# multiproxy-go

A multi-protocol proxy server written in Go supporting **HTTP, SOCKS4, SOCKS5, and Minecraft proxying** — all running on a single port.

---

## 🚀 Usage

### 🔧 Build it yourself

```bash
git clone https://github.com/adithya5434/multiproxy-go.git
cd multiproxy-go

go build -o multiproxy .
./multiproxy
```

---

### 📦 Run from releases

1. [Download latest release](https://github.com/adithya5434/multiproxy-go/releases/latest) for your OS (Windows/Linux)

2. Run:

**Linux**

```bash
chmod +x multiproxy-linux-amd64
./multiproxy-linux-amd64
```

**Windows**

```powershell
.\multiproxy-windows-amd64.exe
```

---

## ⚙️ Configuration

On first run, if no config file is found:

```
Config not found. Creating default config: proxy_config.json
```

A default config will be generated.
Edit this file and restart the proxy.

---

## 🌐 Supported Protocols

* HTTP Proxy (with CONNECT support)
* HTTP Web Proxy
* SOCKS4 / SOCKS4a
* SOCKS5 (optional authentication)
* Minecraft TCP proxy

---

## 🧠 How it works

* Listens on a single configured IP/port
* Detects protocol automatically (HTTP / SOCKS4 / SOCKS5 / MC)
* Routes connection to the appropriate handler
* Relays traffic between client and target

---

## 🧪 Testing

Example using SOCKS5:

```bash
curl --proxy socks5://127.0.0.1:1080 http://example.com
```

You can also test using:

* `proxy_test.py`
* browser proxy settings
* tools like proxychains

---

## 🛠️ TODO

* [ ] Authentication in  HTTP proxy 
* [X] Add web proxy support

