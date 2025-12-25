# GHOST_PACKET

🚀 **GHOST_PACKET** is a **Simple yet Powerful DOS Load Testing Tool** written in **Go**, designed for **educational purposes** and **authorized security testing only**.

---

## ⚠️ Warning & Disclaimer

This tool is intended **only for educational purposes** and **authorized security testing**.  
Unauthorized usage on systems you do **not own** or do **not have explicit permission to test** is **illegal** and can result in **criminal charges**.

By using this tool, you **assume full responsibility** for any consequences, including **data loss**, **service disruption**, or **legal issues**.

---

## Features

- Auto HTTP method detection (GET, POST, HEAD, etc.)  
- Supports custom JSON payloads for POST/PUT/PATCH requests  
- Randomized headers and IPs for more realistic traffic simulation  
- Rate-limited requests (RPS) with configurable workers  
- Real-time status code reporting with emojis  
- Graceful shutdown on Ctrl+C or system signals  
- Supports HTTPS with self-signed certificates  

---

## Installation

### Requirements

- [Go 1.20+](https://golang.org/dl/) installed  
- Works on **Windows, Linux, macOS**  

### Build

```bash
git clone https://github.com/<yourusername>/ghost_packet.git
cd ghost_packet
go build -o ghost_packet main.go
```

### Run

```bash
./ghost_packet
```

---

## Usage

1. Launch tool:

```bash
./ghost_packet
```

2. Enter target URL:

```
Target URL (https://example.com):
```

3. Choose HTTP Method (optional, leave blank for auto-detect)  
4. Enter JSON Body if needed (optional for POST/PUT/PATCH)  
5. Configure test parameters:

```
Duration (seconds): 60
Workers: 10
RPS limit: 50
```

6. Start load test. Output example:

```
[✅ OK] https://example.com
[⚠️ Forbidden] https://example.com
[☠️ Server Error] https://example.com
```

Summary after test:

```
✅ Test completed
📊 Total: 500 | Success: 480 | Unreachable: 20
```

---

## Status Codes & Emojis

| Code | Meaning |
|------|---------|
| 200  | ✅ OK |
| 201  | 🎉 Created |
| 202  | ⏳ Accepted |
| 204  | 🈳 No Content |
| 301  | 📍 Moved Permanently |
| 302  | ↩️ Found |
| 400  | ❌ Bad Request |
| 401  | 🔒 Unauthorized |
| 403  | ⚠️ Forbidden |
| 404  | 🛑 Not Found |
| 408  | ⏱️ Request Timeout |
| 429  | 🚨 Too Many Requests |
| 500  | ☠️ Server Error |
| 503  | 🛑 Service Unavailable |
| 504  | ⏳ Gateway Timeout |

*(Full list included in source code)*

---

## Configuration Options

| Option | Description |
|--------|-------------|
| Target URL | URL to test (HTTPS or HTTP) |
| HTTP Method | GET, POST, HEAD, PUT, PATCH (auto-detect available) |
| JSON Body | Optional JSON payload for POST/PUT/PATCH |
| Duration | Test duration in seconds |
| Workers | Number of concurrent workers |
| RPS Limit | Maximum requests per second |

---

## Contributing

Contributions are welcome for:  

- Bug fixes  
- Feature enhancements  
- Code optimizations  

Please **fork the repository**, make your changes, and submit a **pull request**.

---

## License

This tool is intended for **educational purposes only**.  
Use responsibly and **only on systems you own or are authorized to test**.  
No liability is assumed by the author or organization.

---

## Author

**Z4psT3r**  
Organization: **HonkSec**  

---

## Support

If you encounter issues, open an **issue** on GitHub or contact the author.
