Hash Cracker v1.0

High-performance password hash recovery tool written in Go, supporting parallel processing.
⚠️ Disclaimer

For legal use only:
Intended for recovering your own passwords, security testing with explicit owner permission, and educational purposes.
Unauthorized use for accessing third-party systems is strictly prohibited.
🚀 Features

    Dictionary attacks: Handles wordlists with over 2M+ passwords

    Automatic mutations: Generates up to 35 variants per password candidate

    Parallel processing: Utilizes all CPU cores with Go goroutines

    Supported algorithms: MD5, SHA-1, SHA-256, SHA-512

    High speed: Up to 5.6M passwords/sec (SHA-512)

🛠️ Installation

bash
git clone https://github.com/porgnope/hash-cracker.git
cd hash-cracker
go build -o hash_cracker

🔧 Technologies

Why Go instead of Python?

    Goroutines: Lightweight (2KB vs 1–2MB for Python threads)

    No GIL: Real parallelism across all CPU cores

    Native compilation: Direct machine code execution

    Performance: Typically 3–5× faster than Python multiprocessing

📋 Requirements

    Go 1.16+

    512 MB RAM (2+ GB recommended)

    Multi-core CPU (optimized for parallelism)

📜 License

MIT License—free use with attribution required.

Engineered to demonstrate Go’s parallel processing and cryptographic capabilities.
