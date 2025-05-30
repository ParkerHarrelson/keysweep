keysweep/
├── scanner-cli/                # Go module
│   ├── main.go
│   └── go.mod
├── control-plane/              # Kotlin + Spring Boot
│   ├── src/main/kotlin/…
│   └── build.gradle.kts
├── pr-bot/                     # Java 22 (JGit)
│   └── src/…
├── detectors/                  # future plugin JARs
├── rotators/                   # plugin SPI impls
├── action/                     # GitHub Action Docker context
│   ├── Dockerfile
│   └── wrapper.sh
├── charts/                     # Helm (enterprise)
├── terraform/                  # IaC examples
├── .gitleaks.toml              # default ruleset
└── docs/
    └── ARCHITECTURE.md
