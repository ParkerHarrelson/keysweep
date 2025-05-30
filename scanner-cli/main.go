package main

import (
	"bufio"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema/v1/keysweep-finding.schema.json
var findingSchema string

type gitleaksFinding struct {
	RuleID    string `json:"RuleID"`
	File      string `json:"File"`
	StartLine int    `json:"StartLine"`
	Secret    string `json:"Secret"`
	Commit    string `json:"Commit"`
}

type Finding struct {
	RuleID string `json:"rule_id"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Secret string `json:"secret"`
	Commit string `json:"commit"`
}

var debug = os.Getenv("KEYSWEEP_DEBUG") != ""

func dprintf(format string, a ...any) {
	if debug {
		log.Printf(format, a...)
	}
}

func main() {
	rawDiff, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("read stdin: %v", err)
	}
	dprintf("stdin bytes: %d", len(rawDiff))

	var filtered bytes.Buffer
	sc := bufio.NewScanner(bytes.NewReader(rawDiff))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "diff ") ||
			strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "--- ") ||
			strings.HasPrefix(line, "+++ ") ||
			strings.HasPrefix(line, "@@ ") {
			continue
		}
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "++") {
			filtered.WriteString(line[1:])
			filtered.WriteByte('\n')
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("scan diff: %v", err)
	}
	if filtered.Len() == 0 {
		dprintf("no added-line chunks detected – falling back to full input")
		filtered.Write(rawDiff)
	}
	dprintf("filtered bytes: %d", filtered.Len())

	cmd := exec.Command(
		"/bin/gitleaks", "stdin",
		"--config", "/workspace/gitleaks.toml",
		"--report-format", "json",
		"--report-path", "-",
		"--no-banner", "--log-level", "fatal",
		"--exit-code", "0",
	)
	cmd.Stdin = bytes.NewReader(filtered.Bytes())

	out, err := cmd.CombinedOutput()
	dprintf("gitleaks exit error: %v", err)
	dprintf("gitleaks combined output:\n%s", string(out))
	if err != nil && !acceptable(err) {
		log.Fatalf("gitleaks exec failed: %v\nraw output:\n%s", err, string(out))
	}

	findings, err := parse(out)
	if err != nil {
		log.Fatalf("decode findings: %v", err)
	}
	dprintf("parsed findings: %+v", findings)

	if len(findings) == 0 {
		log.Println("✅  no secrets found")
		return
	}

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	payload, _ := json.Marshal(findings)
	if err := validateAgainstSchema(payload); err != nil {
		log.Fatalf("schema validation failed: %v", err)
	}
	sig := ed25519.Sign(priv, payload)
	if url := os.Getenv("KEYSWEEP_URL"); url != "" {
		req, _ := http.NewRequest("POST", url+"/findings", bytes.NewReader(payload))
		req.Header.Set("X-Signature-Ed25519", hex.EncodeToString(sig))
		req.Header.Set("X-Public-Key", hex.EncodeToString(pub))
		http.DefaultClient.Do(req)
	}

	log.Printf("❌  %d secret(s) found – failing build", len(findings))
	os.Exit(1)
}

func acceptable(err error) bool {
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode() == 1
	}
	return false
}

func parse(out []byte) ([]Finding, error) {
	var glfs []gitleaksFinding
	trim := bytes.TrimSpace(out)
	if len(trim) > 0 && trim[0] == '[' {
		if err := json.Unmarshal(trim, &glfs); err != nil {
			return nil, fmt.Errorf("unmarshal array: %w", err)
		}
	} else {
		dec := json.NewDecoder(bytes.NewReader(out))
		for {
			var gf gitleaksFinding
			if err := dec.Decode(&gf); err != nil {
				if err == io.EOF {
					break
				}
				return nil, fmt.Errorf("decode ndjson: %w", err)
			}
			if gf.RuleID == "" {
				continue
			}
			glfs = append(glfs, gf)
		}
	}

	var res []Finding
	for _, gf := range glfs {
		line := gf.StartLine
		if line < 1 {
			line = 1
		}
		res = append(res, Finding{
			RuleID: gf.RuleID,
			File:   gf.File,
			Line:   line,
			Secret: gf.Secret,
			Commit: gf.Commit,
		})
	}

	return res, nil
}

func validateAgainstSchema(doc []byte) error {
	arraySchema := fmt.Sprintf(`{"type":"array","items":%s}`, findingSchema)
	schemaLoader := gojsonschema.NewStringLoader(arraySchema)
	docLoader := gojsonschema.NewBytesLoader(doc)

	res, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return err
	}
	if !res.Valid() {
		var all []string
		for _, e := range res.Errors() {
			all = append(all, e.String())
		}
		return fmt.Errorf("schema errors: %v", all)
	}
	return nil
}
