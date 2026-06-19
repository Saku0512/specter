package scenario

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  specter scenario [flags] <name>
  specter scenario [flags] list
  specter scenario [flags] current

Flags:
  --url <url>    Specter server URL (default: http://localhost:8080)

Examples:
  specter scenario login-success
  specter scenario --url http://localhost:3000 list

`)
}

func Run(args []string) {
	fs := flag.NewFlagSet("scenario", flag.ExitOnError)
	fs.Usage = usage
	baseURL := fs.String("url", "http://localhost:8080", "Specter server URL")
	if val := os.Getenv("SPECTER_URL"); val != "" {
		*baseURL = val
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	rest := fs.Args()
	if len(rest) == 0 || rest[0] == "current" {
		if err := printCurrent(*baseURL); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if rest[0] == "list" {
		if err := listScenarios(*baseURL); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(rest) > 1 {
		usage()
		os.Exit(2)
	}
	if err := applyScenario(*baseURL, rest[0]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type scenariosResponse struct {
	Active    string   `json:"active"`
	Scenarios []string `json:"scenarios"`
}

func endpoint(baseURL, path string) string {
	return strings.TrimRight(baseURL, "/") + path
}

func decodeError(resp *http.Response) string {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.Error != "" {
		return payload.Error
	}
	if len(body) > 0 {
		return string(body)
	}
	return resp.Status
}

func fetchScenarios(baseURL string) (scenariosResponse, error) {
	resp, err := http.Get(endpoint(baseURL, "/__specter/scenarios"))
	if err != nil {
		return scenariosResponse{}, fmt.Errorf("failed to connect to specter at %s: %w", baseURL, err)
	}
	if resp.StatusCode != http.StatusOK {
		return scenariosResponse{}, fmt.Errorf("failed to list scenarios: %s", decodeError(resp))
	}
	defer resp.Body.Close()
	var payload scenariosResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return scenariosResponse{}, fmt.Errorf("failed to decode scenarios response: %w", err)
	}
	return payload, nil
}

func printCurrent(baseURL string) error {
	payload, err := fetchScenarios(baseURL)
	if err != nil {
		return err
	}
	if payload.Active == "" {
		fmt.Println("active scenario: (none)")
	} else {
		fmt.Printf("active scenario: %s\n", payload.Active)
	}
	if len(payload.Scenarios) == 0 {
		fmt.Println("available scenarios: (none)")
		return nil
	}
	fmt.Println("available scenarios:")
	for _, name := range payload.Scenarios {
		fmt.Printf("  %s\n", name)
	}
	return nil
}

func listScenarios(baseURL string) error {
	payload, err := fetchScenarios(baseURL)
	if err != nil {
		return err
	}
	if len(payload.Scenarios) == 0 {
		fmt.Println("no scenarios configured")
		return nil
	}
	for _, name := range payload.Scenarios {
		if name == payload.Active {
			fmt.Printf("* %s\n", name)
		} else {
			fmt.Printf("  %s\n", name)
		}
	}
	return nil
}

func applyScenario(baseURL, name string) error {
	req, err := http.NewRequest(http.MethodPost, endpoint(baseURL, "/__specter/scenarios/"+name), bytes.NewReader(nil))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to specter at %s: %w", baseURL, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to apply scenario %q: %s", name, decodeError(resp))
	}
	defer resp.Body.Close()
	fmt.Printf("applied scenario: %s\n", name)
	return nil
}
